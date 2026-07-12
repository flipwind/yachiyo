package prompt

import (
	"fmt"
	"os"
	"time"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Context struct {
	History history.HistoryStorage
	State   state.State
	Emotion state.Emotion
	Factors initiative.Factors
	Note string
}

func UserPromptBuilder(c *Context, t *trigger.Message) []history.History {
	if len(c.History.ListAll()) == 0 {
		systemPrompt, err := os.ReadFile("../yachiyo-runtime/assets/systemPrompt.md")
		if err != nil {
		}

		c.History.Remember(history.History{
			Role:    "system",
			Content: string(systemPrompt),
			Time: time.Now(),
		})
	}

	// Current Message Build
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`[UserMessage]
For This Conversation ONLY:
This part, either Emotion and State, you must follow it, in this round of conversation.
<Yachiyo Runtime>
Time: %s
Emotion: %s
State: %s
Last Conversation Active: %s
Session Context: %s
---
User Content: %s
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.
`,
		currentTime, c.Emotion.String(), c.State.Prompt(), c.History.GetLastActive(), c.Note, t.String())

	c.History.Remember(history.History{
		Role:    "user",
		Content: prompt,
		Time: time.Now(),
	})

	ylog.Debug("Prompt built: %s", prompt)

	return c.History.ListAll()
}

func InitiativePromptBuilder(c *Context) []history.History {
	if len(c.History.ListAll()) == 0 {
		ylog.Error("History is empty.")
		return nil
	}

	// Current Message Build
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`[InitiativeMessage]
For This Conversation ONLY:
This part, either Emotion and State, you must follow it, in this round of conversation.
---
<Yachiyo Runtime>
Time: %s
Emotion: %s
State: %s
Last Conversation Active: %s
Session Context: %s
---
<Runtime Factors>
%s
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.
`,
		currentTime, c.Emotion.String(), c.State.Prompt(), c.History.GetLastActive().Format("2006.01.02 15:04:05"), c.Note, c.Factors.String())

	c.History.Remember(history.History{
		Role:    "user",
		Content: prompt,
		Time: time.Now(),
	})

	ylog.Debug("Prompt built: %s", prompt)

	return c.History.ListAll()
}
