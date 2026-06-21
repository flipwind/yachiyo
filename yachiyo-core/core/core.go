package core

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/llm/fake"
	"yachiyo/yachiyo-core/memory"
	"yachiyo/yachiyo-core/memory/basic"
)

type Core struct{
	Memory memory.MemoryStorage
}

func New() *Core{
	return &Core{
		Memory: basic.New(),
	}
}

func (c *Core) Process(e *event.UserMessageEvent) string{
	c.Memory.Remember(memory.Memory{
		Content: e.Message.String(),
	})

	systemPrompt := "systemPrompt"
	currentTime := time.Now().Format("2006.01.02 15:04:05")
	currentEmotion := "Excited"
	currentState := "Socialization: High"

	prompt := fmt.Sprintf(`%s
Time: %s
Emotion: %s
State: <%s>
Context: %s`, 
systemPrompt, currentTime, currentEmotion, currentState, c.Memory.ListAll())

	llm := fake.NewFakeLLM()

	result := llm.Gen(prompt)
	
	return result
}