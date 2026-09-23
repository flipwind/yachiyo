package core

import (
	"fmt"
	"slices"
	"sync"
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/history/basic"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/yerror"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	History       history.HistoryStorage
	State         state.State
	Emotion       state.Emotion
	LLM           llm.LLM
	Config        config.Config
	Determination state.Determination
	Factors       initiative.Factors

	JSONConstraint bool
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
		return nil, yerror.TypeMissing("Config")
	}

	// TODO: LLM List
	llmConfig := config.LLM.DefaultProvider
	llm := provider.NewOpenAIProvider(*llmConfig.BaseUrl, *llmConfig.Secret, *llmConfig.Model)

	initiativeConfig := config.Initiative

	factors, errs := initiative.NewFactors(
		// TODO: reflect
		*initiativeConfig.Threshold,

		initiativeConfig.Factors.Sociability,
		initiativeConfig.Factors.AloneTime,
		initiativeConfig.Factors.Daytime,
	)

	if len(errs) != 0 {
		for _, err := range errs {
			ylog.Error("Factor curve error: %v", err)
		}
		return nil, yerror.FieldInvalid("Factor", "curves invalid")
	}

	core := &Core{
		History:        basic.New(),
		State:          state.NewState(),
		Emotion:        state.NewEmotion(),
		LLM:            llm,
		Config:         config,
		Determination:  state.NewDetermination(),
		Factors:        factors,
		JSONConstraint: false,
		Note:           "",
		LastActiveTime: time.Now(),
	}

	core.Pipe = NewPipeline(core.Process)

	return core, nil
}

func (c *Core) Run() {
	go c.Pipe.Listen()
	go c.Pipe.DecisionListen()
	go c.Pipe.DistributionListen()
	go c.Clock()
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
	msg += fmt.Sprintf("factors: %v\n", c.Factors.String())

	return msg
}

func (c *Core) AppendHistory(hist history.History) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.History.Append(hist)
}

// Core snapshot

type Snapshot struct {
	State          state.State
	Emotion        state.Emotion
	Determination  state.Determination
	Factors        initiative.Factors
	JSONConstraint bool
	Note           string
	LastActiveTime time.Time
}


func (c *Core) snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := Snapshot{
		State:          c.State,
		Emotion:        c.Emotion,
		Determination:  c.Determination,
		Factors:        c.Factors,
		JSONConstraint: c.JSONConstraint,
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

func GetLastUserAddress(hist []history.History) (address.Address, bool) {
	for _, h := range slices.Backward(hist) {
		if item, ok := h.(history.UserMessage); ok {
			return item.Address, true
		}
	}

	return address.Address{}, false
}
