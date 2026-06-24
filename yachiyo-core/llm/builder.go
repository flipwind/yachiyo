package llm

import (
	"fmt"
	"os"
	"time"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/history"
	"yachiyo/yachiyo-core/state"
)

type Context struct {
	History history.HistoryStorage
	State   state.State
	Emotion state.Emotion
}

func PromptBuilder(c *Context, e *event.UserMessageEvent) []history.History {
	if len(c.History.ListAll()) == 0 {
		systemPrompt, err := os.ReadFile("../yachiyo-core/assets/systemPrompt.md")
		if err != nil {
		}

		c.History.Remember(history.History{
			Role:    "system",
			Content: string(systemPrompt),
		})
	}

	// Current Message Build
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`Time: %s
Emotion: %s
State: %s
Content: %s`,
		currentTime, c.Emotion.String(), c.State.Prompt(), e.Message.String())

	c.History.Remember(history.History{
		Role:    "user",
		Content: prompt,
	})

	return c.History.ListAll()
}
