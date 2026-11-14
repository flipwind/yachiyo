package core

import (
	"fmt"
	"sync"
	"time"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/history/sqlite3"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	History      history.HistoryStorage
	compressNode history.CompressNode
	LLM          llm.Client
	Config       config.Config
	State        *state.State

	Note         string
	LLMBusy      bool
	CompressBusy bool

	Pipe *Pipeline  //exhaustruct:optional
	mu   sync.Mutex //exhaustruct:optional
}

func New() (*Core, error) {
	// TODO: change config status (dev)
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("Core loading config failed: %w", err)
	}

	// TODO: LLM List
	llmConfig := config.LLM.DefaultProvider
	llm := provider.NewOpenAIProvider(*llmConfig.BaseUrl, *llmConfig.Secret, *llmConfig.Model)

	initiativeConfig := config.Initiative
	personality := state.BigFive{
		Openness:          config.Core.BigFive.Openness,
		Conscientiousness: config.Core.BigFive.Conscientiousness,
		Extraversion:      config.Core.BigFive.Extraversion,
		Agreeableness:     config.Core.BigFive.Agreeableness,
		Neuroticism:       config.Core.BigFive.Neuroticism,
	}

	storage, restoreResult, err := sqlite3.New()
	if err != nil {
		return nil, fmt.Errorf("Storage created error: %w", err)
	}

	// Restore state
	checkpoint := restoreResult.Checkpoint
	checkpointExist := checkpoint.Exist

	ylog.Success("Now history sequence at %v.", checkpoint.NowSequence)

	var core *Core

	energy, err := state.NewEnergy(*initiativeConfig.Energy.Curve)
	if err != nil {
		return nil, fmt.Errorf("Core loading with invalid curves: %w", err)
	}

	snap := state.Initial(personality, time.Now())
	elapsed := time.Duration(0)
	if checkpointExist {
		snap = checkpoint.State
		elapsed = time.Since(checkpoint.StateAt)
	}
	st := state.Restore(personality, energy, snap)
	if err := st.Advance(elapsed, time.Now()); err != nil {
		return nil, fmt.Errorf("State advanced error: %w", err)
	}

	core = &Core{
		History:      storage,
		compressNode: restoreResult.LatestNode,
		LLM:          llm,
		Config:       config,
		State:        st,
		Note:         checkpoint.Note,
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
	c.History.Close()
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

func (c *Core) AppendHistory(hist event.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.History.Append(hist); err != nil {
		ylog.Error("Append history error: %v", err)
		return err
	}
	return nil
}

func (c *Core) commitCompressNode(node history.CompressNode) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.History.SaveCompressNode(node); err != nil {
		ylog.Error("Append compressnode error: %v", err)
		return err
	}
	c.compressNode = node

	return nil
}

func (c *Core) getHistory() []event.Event {
	c.mu.Lock()
	defer c.mu.Unlock()

	hist := c.History.ListAll()

	return hist
}

func (c *Core) getSystemPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Config.Prompt.SystemPrompt
}

func (c *Core) getCompressPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Config.Prompt.CompressPrompt
}

func (c *Core) getNote() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Note
}

func (c *Core) getContextWindow() history.ContextWindow {
	c.mu.Lock()
	defer c.mu.Unlock()

	hist := c.History.ListAll()
	if !c.compressNode.Empty {
		hist = c.History.ListAfter(c.compressNode.CompressEventIDTo)
	}

	return history.ContextWindow{
		CurrentNode:      c.compressNode,
		CurrentHistories: hist,
	}
}
