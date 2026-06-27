package core

import (
	"yachiyo/yachiyo-core/config"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/history"
	"yachiyo/yachiyo-core/history/basic"
	"yachiyo/yachiyo-core/llm"
	"yachiyo/yachiyo-core/llm/provider"
	"yachiyo/yachiyo-core/prompt"
	"yachiyo/yachiyo-core/state"
	"yachiyo/yachiyo-utils/logger"
)

var log = logger.New("Yachiyo.Core")

type Core struct{
	History history.HistoryStorage
	State state.State
	Emotion state.Emotion
	LLM llm.LLM
	Config config.ConfigManager
}

func New() *Core{
	config, err := config.LoadConfig("../yachiyo-core/assets/config.yaml")
	if err != nil {}

	// TODO: LLM List
	llmConfig := config.CurrentConfig.LLM.Key[0]
	llm := provider.NewOpenAIProvider(llmConfig.BaseUrl, llmConfig.Secret, llmConfig.ModelName)

	return &Core{
		History: basic.New(),
		State: state.NewState(),
		Emotion: state.NewEmotion(),
		LLM: llm,
		Config: config,
	}
}

func (c *Core) Process(e *event.UserMessageEvent) string{
	histories := prompt.PromptBuilder(&prompt.Context{
		History: c.History,
		Emotion: c.Emotion,
		State: c.State,
	}, e)

	result, err := c.LLM.Gen(histories)
	
	if err != nil {
		log.Error("LLM Generating error: %v", err)
	}

	answer := c.OutputProcess(result)

	c.History.Remember(history.History{
		Role:    "assistant",
		Content: answer,
	})

	return result
}