package prompt

import (
	"fmt"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Context struct {
	SystemPrompt string
	History      []history.History
	Note         string
}

func StatementPrompt(snap history.Snapshot, note string) SystemPrompt {
	content := fmt.Sprintf(`<Yachiyo Runtime>
In this round of conversation, you must consider these statements and follow them.
Emotion: %s
State: %s
Last Active Time: %s
Session Context: %s
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.`,
		snap.Emotion.String(),
		snap.State.Prompt(),
		snap.LastActiveTime.Format("2006.01.02 15:04:05"),
		note,
	)

	return SystemPrompt{
		Content: content,
	}
}

func InitiativePrompt(snap history.Snapshot) SystemPrompt {
	content := fmt.Sprintf(`[Initiative Trigger]
<Runtime Factors>
Factors determined whether to active initiative message. As the result, when you read this, it means the initiative threshold was reached.
Now the factors are given to know why you should send initiative message.
Following are some percentage. Notice that percentage is accumulated with the time normally.
%s`,
		snap.Factors.String(),
	)

	return SystemPrompt{
		Content: content,
	}
}

func HistoryPrompt(histories []history.History) []Prompts {
	var prompts []Prompts
	for _, h := range histories {
		switch hist := h.(type) {
		case history.UserMessage:
			prompts = append(prompts, UserPrompt{
				Content: hist.Render(),
			})
		case history.AssistantMessage:
			prompts = append(prompts, AssistantPrompt{
				Content: hist.Render(),
			})
		default:
			ylog.Debug("Unsupport history: %T", hist)
		}
	}

	return prompts
}

func UserPromptBuilder(c Context, t *trigger.Message, snap history.Snapshot) []Prompts {
	var prompts []Prompts

	// 1. System Prompt
	prompts = append(prompts, SystemPrompt{Content: c.SystemPrompt})

	// 2. History Prompt (Including new message)
	prompts = append(prompts, HistoryPrompt(c.History)...)

	// 3. Current Statement
	statementPrompt := StatementPrompt(snap, c.Note)
	prompts = append(prompts, statementPrompt)

	ylog.Info("Received user message [%v]", t.String())

	return prompts
}

func InitiativePromptBuilder(c Context, snap history.Snapshot) []Prompts {
	// 0. Check if the trigger vaild
	if len(c.History) == 0 {
		ylog.Error("History is empty.")
		return nil
	}

	// 1. System Prompt
	var prompts []Prompts
	prompts = append(prompts, SystemPrompt{Content: c.SystemPrompt})

	// 2. History Prompt (Including new message)
	prompts = append(prompts, HistoryPrompt(c.History)...)

	// 3. Current Statement
	statementPrompt := StatementPrompt(snap, c.Note)
	prompts = append(prompts, statementPrompt)

	// 4. Initiative trigger reason
	initiativePrompt := InitiativePrompt(snap)
	prompts = append(prompts, initiativePrompt)

	ylog.Info("Initiative Message triggered.")

	return prompts
}
