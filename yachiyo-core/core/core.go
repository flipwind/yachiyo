package core

import (
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/llm"
	"yachiyo/yachiyo-core/llm/fake"
	"yachiyo/yachiyo-core/memory"
	"yachiyo/yachiyo-core/memory/basic"
	"yachiyo/yachiyo-core/state"
)

type Core struct{
	Memory memory.MemoryStorage
	State state.State
	Emotion state.Emotion
	LLM llm.LLM
}

func New() *Core{
	return &Core{
		Memory: basic.New(),
		State: state.NewState(),
		Emotion: state.NewEmotion(),
		LLM: fake.NewFakeLLM(),
	}
}

func (c *Core) Process(e *event.UserMessageEvent) string{
	c.Memory.Remember(memory.Memory{
		Content: e.Message.String(),
	})

	prompt := llm.PromptBuilder(&llm.Context{
		Memory: c.Memory,
		Emotion: c.Emotion,
		State: c.State,
	})

	result := c.LLM.Gen(prompt)
	
	return result
}