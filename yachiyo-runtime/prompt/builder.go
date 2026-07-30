package prompt

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Context struct {
	SystemPrompt   string
	History        history.HistoryStorage
	State          state.State
	Emotion        state.Emotion
	Factors        initiative.Factors
	Note           string
	LastActiveTime time.Time
}

func UserPromptBuilder(c *Context, t *trigger.Message) []history.History {
	if len(c.History.ListAll()) == 0 {
		c.History.Remember(history.History{
			Role:    "system",
			Content: string(c.SystemPrompt),
			Time:    time.Now(),
		})
	}

	// Current Message Build
	currentTime := time.Now().Format("2006.01.02 15:04:05")

	prompt := fmt.Sprintf(`[UserMessage]
<Yachiyo Runtime>
This part, either Emotion and State, you must follow it, in this round of conversation.
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
		currentTime, c.Emotion.String(), c.State.Prompt(), c.LastActiveTime.Format("2006.01.02 15:04:05"), c.Note, t.String())

	c.History.Remember(history.History{
		Role:    "user",
		Content: prompt,
		Time:    time.Now(),
		Address: t.Address,
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
---
<Yachiyo Runtime>
This part, either Emotion and State, you must follow it, in this round of conversation.
Time: %s
Emotion: %s
State: %s
Last Conversation Active: %s
Session Context: %s
---
<Runtime Factors>
Factors will active initiative message. As the result, these factors are given to know why you should send initiative message.
Following are some percentage. Notice that percentage is accumulated with the time normally.
%s
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.
`,
		currentTime, c.Emotion.String(), c.State.Prompt(), c.LastActiveTime.Format("2006.01.02 15:04:05"), c.Note, c.Factors.String())

	c.History.Remember(history.History{
		Role:    "user/runtime",
		Content: prompt,
		Time:    time.Now(),
	})

	ylog.Debug("Prompt built: %s", prompt)

	return c.History.ListAll()
}
