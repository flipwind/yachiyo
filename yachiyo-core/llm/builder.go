package llm

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-core/memory"
	"yachiyo/yachiyo-core/state"
)

type Context struct {
	Memory memory.MemoryStorage
	State state.State
	Emotion state.Emotion
}

func PromptBuilder(c *Context) string {
	systemPrompt := "systemPrompt"
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`%s
Time: %s
Emotion: %s
State: %s
Context: %s`,
		systemPrompt, currentTime, c.Emotion.String(), c.State.String(), c.Memory.ListAll())

	return prompt
}
