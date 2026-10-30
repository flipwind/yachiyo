package core

import (
	"fmt"
	"sync"
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/history/sqlite3"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	History       history.HistoryStorage
	State         state.State
	LLM           llm.LLM
	Config        config.Config
	Determination state.Determination
	Energy        initiative.Energy
	Affect        state.Engine

	Note           string
	Pipe           *Pipeline
	LastActiveTime time.Time
	LLMBusy        bool

	mu sync.Mutex
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

	storage, err := sqlite3.New()
	if err != nil {
		return nil, fmt.Errorf("Storage created error: %w", err)
	}

	// Restore state
	checkpoint := storage.Checkpoint
	checkpointExist := checkpoint.Exist

	var core *Core

	if checkpointExist {
		affect := state.NewEngine(state.BigFive{
			Openness:          config.Core.BigFive.Openness,
			Conscientiousness: config.Core.BigFive.Conscientiousness,
			Extraversion:      config.Core.BigFive.Extraversion,
			Agreeableness:     config.Core.BigFive.Agreeableness,
			Neuroticism:       config.Core.BigFive.Neuroticism,
		}, &checkpoint.Affect.Current.Position, &checkpoint.Affect.Emotions)

		energy, err := initiative.NewEnergy(*initiativeConfig.Energy.Curve, &checkpoint.Energy.Value)

		if err != nil {
			return nil, fmt.Errorf("Core loading with invalid curves: %w", err)
		}

		core = &Core{
			History:       storage,
			State:         checkpoint.State,
			LLM:           llm,
			Config:        config,
			Determination: checkpoint.Determination,
			Energy:        energy,
			Affect:        affect,

			Note:           checkpoint.Note,
			LastActiveTime: checkpoint.LastActiveTime,
		}

		core.processTimetick(time.Since(checkpoint.StateAt))
	} else {
		affect := state.NewEngine(state.BigFive{
			Openness:          config.Core.BigFive.Openness,
			Conscientiousness: config.Core.BigFive.Conscientiousness,
			Extraversion:      config.Core.BigFive.Extraversion,
			Agreeableness:     config.Core.BigFive.Agreeableness,
			Neuroticism:       config.Core.BigFive.Neuroticism,
		}, nil, nil)

		energy, err := initiative.NewEnergy(*initiativeConfig.Energy.Curve, nil)

		if err != nil {
			return nil, fmt.Errorf("Core loading with invalid curves: %w", err)
		}

		core = &Core{
			History:       storage,
			State:         state.NewState(),
			LLM:           llm,
			Config:        config,
			Determination: state.NewDetermination(),
			Energy:        energy,
			Affect:        affect,

			Note:           "",
			LastActiveTime: time.Now(),
		}
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

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.State.Drives() {
		msg += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	msg += fmt.Sprintf("Energy: %v\n", c.Energy.Value)

	return msg
}

func (c *Core) AppendHistory(hist history.History) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.History.Append(hist); err != nil {
		ylog.Error("Append history error: %v", err)
		return err
	}
	return nil
}

// Core snapshot

type Snapshot struct {
	State          state.State
	Affect         state.Engine
	Determination  state.Determination
	Energy         initiative.Energy
	JSONConstraint bool
	Note           string
	LastActiveTime time.Time
}

func (c *Core) snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := Snapshot{
		State:          c.State,
		Affect:         c.Affect,
		Determination:  c.Determination,
		Energy:         c.Energy,
		Note:           c.Note,
		LastActiveTime: c.LastActiveTime,
	}

	return snap
}

func (c *Core) getHistory() []history.History {
	c.mu.Lock()
	defer c.mu.Unlock()

	hist := append([]history.History(nil), c.History.ListAll()...)

	return hist
}

func (c *Core) getSystemPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Config.Prompt.SystemPrompt
}

func (c *Core) getNote() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Note
}
