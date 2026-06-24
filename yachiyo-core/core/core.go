package core

import (
	"yachiyo/yachiyo-core/config"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/history"
	"yachiyo/yachiyo-core/history/basic"
	"yachiyo/yachiyo-core/llm"
	"yachiyo/yachiyo-core/llm/fake"
	"yachiyo/yachiyo-core/state"
)

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
	return &Core{
		History: basic.New(),
		State: state.NewState(),
		Emotion: state.NewEmotion(),
		LLM: fake.NewFakeLLM(),
		Config: config,
	}
}

func (c *Core) Process(e *event.UserMessageEvent) string{
	history := llm.PromptBuilder(&llm.Context{
		History: c.History,
		Emotion: c.Emotion,
		State: c.State,
	}, e)

	result := c.LLM.Gen(history)
	
	return result
}