package core

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/action"
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
	LastActiveTime *time.Time
	LLMBusy        bool
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

	core := &Core{
		History:        basic.New(),
		State:          state.NewState(),
		Emotion:        state.NewEmotion(),
		LLM:            llm,
		Config:         config,
		Determination:  state.NewDetermination(),
		Factors:        initiative.NewFactors(
			// TODO: reflect
			*initiativeConfig.Threshold,

			initiativeConfig.Factors.Sociability,
			initiativeConfig.Factors.AloneTime,
			initiativeConfig.Factors.Daytime,
		),
		JSONConstraint: false,
		Note:           "",
	}

	core.Pipe = NewPipeline(core.Process)

	return core, nil
}

func (c *Core) Run() {
	go c.Pipe.Listen()
	go c.Pipe.DistributionListen()
	go c.Clock()
}

func (c *Core) Dispatch(t trigger.Trigger) {
	c.Pipe.Raw <- t
}

func (c *Core) Distribution(a action.Action) {
	c.Pipe.Distribution <- a
}

func (c *Core) InternalState() string {
	msg := fmt.Sprintf("Received timetick %s\n", time.Now().Format("15:04:05"))
	for _, s := range c.State.Drives() {
		msg += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	msg += fmt.Sprintf("factors: %v\n", c.Factors.String())

	return msg
}
