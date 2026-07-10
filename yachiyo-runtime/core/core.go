package core

import (
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/history/basic"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	History       history.HistoryStorage
	State         state.State
	Emotion       state.Emotion
	LLM           llm.LLM
	Config        config.ConfigManager
	Determination state.Determination

	JSONConstraint bool
	Note           string
	Pipe           *Pipeline
}

func New() *Core {
	config, err := config.LoadConfig("../yachiyo-runtime/assets/config.yaml")
	if err != nil {
	}

	// TODO: LLM List
	llmConfig := config.CurrentConfig.LLM.Key[0]
	llm := provider.NewOpenAIProvider(llmConfig.BaseUrl, llmConfig.Secret, llmConfig.ModelName)

	core := &Core{
		History:        basic.New(),
		State:          state.NewState(),
		Emotion:        state.NewEmotion(),
		LLM:            llm,
		Config:         config,
		Determination:  state.NewDetermination(),
		JSONConstraint: false,
		Note:           "",
	}

	core.Pipe = NewPipeline(core.Process)

	return core
}

func (c *Core) Run(){
	go c.Pipe.Listen()
	go c.Pipe.DistributionListen()
	go c.Clock()
}
