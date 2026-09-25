package prompt

import (
	"fmt"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Context struct {
	SystemPrompt string
	History      []history.History
	Note         string
}

func statementPrompt(snap history.Snapshot, note string) string {
	return fmt.Sprintf(`<Yachiyo Runtime>
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
}

func factorsPrompt(snap history.Snapshot) string {
	return fmt.Sprintf(`<Runtime Factors>
Following are some percentage of factors. These percentage is accumulated with the time normally.
%s`,
		snap.Factors.String(),
	)
}

// System Prompt Builders
func PassivePrompt(snap history.Snapshot, note string) SystemPrompt {
	content := fmt.Sprintf(`<Triggered by [UserMessage]>
%s`,
		statementPrompt(snap, note))

	return SystemPrompt{
		Content: content,
	}
}

func InitiativePrompt(snap history.Snapshot, note string) SystemPrompt {
	content := fmt.Sprintf(`<Triggered by [InitiativeMessage]>
%s
%s`,
		statementPrompt(snap, note),
		factorsPrompt(snap),
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

func UserPromptBuilder(c Context, snap history.Snapshot) []Prompts {
	var prompts []Prompts

	// 1. System Prompt
	prompts = append(prompts, SystemPrompt{Content: c.SystemPrompt})

	// 2. History Prompt (Including new message)
	prompts = append(prompts, HistoryPrompt(c.History)...)

	// 3. Current Statement
	statementPrompt := PassivePrompt(snap, c.Note)
	prompts = append(prompts, statementPrompt)

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

	// 3. Current Statement & Initiative trigger reason
	statementPrompt := InitiativePrompt(snap, c.Note)
	prompts = append(prompts, statementPrompt)

	return prompts
}
