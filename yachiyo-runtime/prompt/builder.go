package prompt

import (
	"fmt"
	"os"
	"time"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/state"
)

type Context struct {
	History history.HistoryStorage
	State   state.State
	Emotion state.Emotion
}

func PromptBuilder(c *Context, t *trigger.Message) []history.History {
	if len(c.History.ListAll()) == 0 {
		systemPrompt, err := os.ReadFile("../yachiyo-runtime/assets/systemPrompt.md")
		if err != nil {
		}

		c.History.Remember(history.History{
			Role:    "system",
			Content: string(systemPrompt),
		})
	}

	// Current Message Build
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`For This Conversation ONLY:
This part, either Emotion and State, you must follow it, in this round of conversation.
<Yachiyo Runtime>
Time: %s
Emotion: %s
State: %s
---
User Content: %s
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.
`,
		currentTime, c.Emotion.String(), c.State.Prompt(), t.String())

	c.History.Remember(history.History{
		Role:    "user",
		Content: prompt,
	})

	return c.History.ListAll()
}
