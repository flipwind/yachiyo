package prompt

import (
	"fmt"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Context struct {
	SystemPrompt  string
	ContextWindow history.ContextWindow
	Note          string
}

func statementPrompt(snap state.Snapshot, note string) string {
	return fmt.Sprintf(`<Yachiyo Runtime>
In this round of conversation, you must consider these statements and follow them.
State: %s
Last Active Time: %s
Session Context: %s
---
<Current Affect Engine>
%s
---
<Energy>
This value is in [0, 1], which the more the value, the more energetic.
%v
---
Whatever the answer is, Remember YOU **MUST** FOLLOW THE JSON OUTPUT RULE.
OUTPUT JSON ONLY. OUTPUT SHOULD ONLY START WITH '{' AND END WITH '}'.`,
		Drives(snap.Drives),
		snap.LastActive.Format("2006.01.02 15:04:05"),
		note,
		Affect(snap.Mood, snap.Emotions),
		snap.Energy,
	)
}

// System Prompt Builders
func PassivePrompt(snap state.Snapshot, note string) llm.Message {
	content := fmt.Sprintf(`<Triggered by [UserMessage]>
%s`,
		statementPrompt(snap, note))

	return llm.Message{
		Role:    llm.RoleSystem,
		Content: content,
	}
}

func InitiativePrompt(snap state.Snapshot, note string) llm.Message {
	content := fmt.Sprintf(`<Triggered by [InitiativeMessage]>
%s`,
		statementPrompt(snap, note),
	)

	return llm.Message{
		Role:    llm.RoleSystem,
		Content: content,
	}
}

func HistoryPrompt(window history.ContextWindow) []llm.Message {
	var msgs []llm.Message

	if window.CurrentNode.Empty == false {
		// Compression Message
		msgs = append(msgs, llm.Message{
			Role:    llm.RoleUser,
			Content: window.CurrentNode.Rend(),
		})
	}

	for _, h := range window.CurrentHistories {
		switch hist := h.(type) {
		case event.UserMessage:
			msgs = append(msgs, llm.Message{
				Role:    llm.RoleUser,
				Content: hist.Render(),
			})
		case event.AssistantMessage:
			msgs = append(msgs, llm.Message{
				Role:    llm.RoleAssistant,
				Content: hist.Render(),
			})
		default:
			ylog.Debug("Unsupport history: %T", hist)
		}
	}

	return msgs
}

func UserPromptBuilder(c Context, snap state.Snapshot) []llm.Message {
	var msgs []llm.Message

	// 1. System Prompt
	msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: c.SystemPrompt})

	// 2. History Prompt (Including new message)
	msgs = append(msgs, HistoryPrompt(c.ContextWindow)...)

	// 3. Current Statement
	statementPrompt := PassivePrompt(snap, c.Note)
	msgs = append(msgs, statementPrompt)

	ylog.Debug("Generated prompt list: %v", msgs)
	return msgs
}

func InitiativePromptBuilder(c Context, snap state.Snapshot) []llm.Message {
	// 0. Check if the trigger vaild
	if c.ContextWindow.IsEmpty() {
		ylog.Error("History is empty.")
		return nil
	}

	// 1. System Prompt
	var msgs []llm.Message
	msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: c.SystemPrompt})

	// 2. History Prompt (Including new message)
	msgs = append(msgs, HistoryPrompt(c.ContextWindow)...)

	// 3. Current Statement & Initiative trigger reason
	statementPrompt := InitiativePrompt(snap, c.Note)
	msgs = append(msgs, statementPrompt)

	ylog.Debug("Generated prompt list: %v", msgs)
	return msgs
}
