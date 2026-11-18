package core

import (
	"fmt"
	"sync"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
	"yachiyo/yachiyo-runtime/store/sqlite3"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	store  store.Store
	conv   *conversation.Conversation
	LLM    llm.Client
	Config config.Config
	State  *state.State

	LLMBusy      bool
	CompressBusy bool

	Pipe *Pipeline  //exhaustruct:optional
	mu   sync.Mutex //exhaustruct:optional
}

func New() (*Core, error) {
	// TODO: change config status (dev)
	cfg, warns, err := config.Load("config.yaml")
	for _, wa := range warns {
		ylog.Warn("Config warning: %v", wa)
	}
	if err != nil {
		return nil, fmt.Errorf("Core loading config failed: %w", err)
	}

	if cfg.LogLevel == config.LogDebug {
		logger.SetLogLevel(logger.Debug)
	} else {
		logger.SetLogLevel(logger.Info)
	}

	// TODO: LLM List
	llmConfig := cfg.LLMProvider
	llm := provider.NewOpenAIProvider(llmConfig.BaseURL, llmConfig.Secret, llmConfig.Model)

	personality := state.BigFive{
		Openness:          cfg.BigFive.Openness,
		Conscientiousness: cfg.BigFive.Conscientiousness,
		Extraversion:      cfg.BigFive.Extraversion,
		Agreeableness:     cfg.BigFive.Agreeableness,
		Neuroticism:       cfg.BigFive.Neuroticism,
	}

	storage, err := sqlite3.Open("yachiyo.db")
	if err != nil {
		return nil, fmt.Errorf("Storage created error: %w", err)
	}

	restored, err := storage.Load()
	if err != nil {
		if err := storage.Close(); err != nil {
			ylog.Error("Storage close error: %v", err)
		}
		return nil, fmt.Errorf("Storage restored error: %w", err)
	}

	var core *Core

	energy, err := state.NewEnergy(cfg.Initiative.EnergyCurve)
	if err != nil {
		return nil, fmt.Errorf("Core loading with invalid curves: %w", err)
	}

	snap := state.Initial(personality, time.Now())
	elapsed := time.Duration(0)
	note := ""
	var sequence int64
	if cp := restored.Checkpoint; cp != nil {
		snap = cp.State
		elapsed = time.Since(cp.SavedAt)
		note = cp.Note
		sequence = cp.Sequence
	}
	
	conv := conversation.Restore(restored.Events, restored.Summary, note, sequence)
	ylog.Success("Now history sequence at %v.", conv.Sequence())
	st := state.Restore(personality, energy, snap)
	if err := st.Advance(elapsed, time.Now()); err != nil {
		return nil, fmt.Errorf("State advanced error: %w", err)
	}

	core = &Core{
		conv:         conv,
		store:        storage,
		LLM:          llm,
		Config:       cfg,
		State:        st,
		LLMBusy:      false,
		CompressBusy: false,
	}

	ylog.Success("Successfully restored/created checkpoint and histories.")

	core.Pipe = NewPipeline(core.Process)

	return core, nil
}

func (c *Core) Run() {
	go c.Pipe.Listen()
	go c.Pipe.DecisionListen()
	go c.Pipe.DistributionListen()
	go c.Clock()
}

func (c *Core) Close() {
	if err := c.store.Close(); err != nil {
		ylog.Error("Storage close error: %v", err)
	}
}

func (c *Core) Dispatch(t trigger.Trigger) {
	select {
	case c.Pipe.Raw <- t:
	default:
		ylog.Warn("dispatch channel full, drop %T", t)
	}
}

func (c *Core) Distribution(a action.Action) {
	select {
	case c.Pipe.Distribution <- a:
	default:
		ylog.Warn("distribution channel full, drop %T", a)
	}
}

func (c *Core) InternalState() string {
	msg := fmt.Sprintf("Received timetick %s\n", time.Now().Format("15:04:05"))

	snap := c.snapshot().State
	for _, s := range snap.Drives.All() {
		msg += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	msg += fmt.Sprintf("Energy: %v\n", snap.Energy)

	return msg
}

func (c *Core) record(e event.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.store.Append(e); err != nil {
		return fmt.Errorf("Storage saving event error: %w", err)
	}

	c.conv.Append(e)
	return nil
}

func (c *Core) commitSummary(summary conversation.Summary) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.store.SaveSummary(summary); err != nil {
		return fmt.Errorf("Storage saving summary error: %w", err)
	}

	c.conv.SetSummary(summary)
	return nil
}

func (c *Core) getSystemPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Config.Prompt.System
}

func (c *Core) getCompressPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Config.Prompt.Compress
}

func (c *Core) getWindow() conversation.Window {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conv.Window()
}

func (c *Core) getEvents() []event.Event {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conv.Events()
}

func (c *Core) getNote() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conv.Note()
}

func (c *Core) setNote(note string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conv.SetNote(note)
}

func (c *Core) newMeta(snap state.Snapshot) event.Meta {
	c.mu.Lock()
	defer c.mu.Unlock()

	return event.Meta{
		ID:        uuid.NewV7(),
		Sequence:  c.conv.NextSequence(),
		Timestamp: time.Now(),
		Snapshot:  snap,
	}
}
