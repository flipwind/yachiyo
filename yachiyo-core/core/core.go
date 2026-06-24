package core

import (
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/llm"
	"yachiyo/yachiyo-core/llm/fake"
	"yachiyo/yachiyo-core/history"
	"yachiyo/yachiyo-core/history/basic"
	"yachiyo/yachiyo-core/state"
)

type Core struct{
	History history.HistoryStorage
	State state.State
	Emotion state.Emotion
	LLM llm.LLM
}

func New() *Core{
	return &Core{
		History: basic.New(),
		State: state.NewState(),
		Emotion: state.NewEmotion(),
		LLM: fake.NewFakeLLM(),
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