package prompt

import (
	"fmt"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Prompt")

type Cause int

const (
	Passive Cause = iota
	Initiative
)

type Input struct {
	Cause        Cause
	SystemPrompt string
	Window       conversation.Window
	State        state.Snapshot
	Note         string
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
		snap.LastActive.Format(timeLayout),
		note,
		Affect(snap.Mood, snap.Emotions),
		snap.Energy,
	)
}

// System Prompt Builders
func (c Cause) trigger() string {
	switch c {
	case Passive:
		return "UserMessage"
	case Initiative:
		return "InitiativeMessage"
	}
	return "Unknown"
}

func HistoryPrompt(window conversation.Window) []llm.Message {
	var msgs []llm.Message

	if window.Summary != nil {
		// Compression Message
		msgs = append(msgs, llm.Message{
			Role:    llm.RoleUser,
			Content: Summary(*window.Summary),
		})
	}

	for _, h := range window.Events {
		switch hist := h.(type) {
		case event.UserMessage:
			msgs = append(msgs, llm.Message{
				Role:    llm.RoleUser,
				Content: UserMessage(hist),
			})
		case event.AssistantMessage:
			msgs = append(msgs, llm.Message{
				Role:    llm.RoleAssistant,
				Content: AssistantMessage(hist),
			})
		default:
			ylog.Error("Unsupport history: %T", hist)
		}
	}

	return msgs
}

func Build(in Input) []llm.Message {
	var msgs []llm.Message

	// 1. System Prompt
	msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: in.SystemPrompt})

	// 2. History Prompt (Including new message)
	msgs = append(msgs, HistoryPrompt(in.Window)...)

	// 3. Current Statement & Initiative trigger reason
	statementPrompt := llm.Message{
		Role: llm.RoleSystem,
		Content: fmt.Sprintf("<Triggered by [%s]>\n%s",
			in.Cause.trigger(),
			statementPrompt(in.State, in.Note),
		),
	}

	msgs = append(msgs, statementPrompt)

	ylog.Debug("Generated prompt list: %v", msgs)
	return msgs
}

func BuildCompression(contextWindow conversation.Window, compressPrompt string) []llm.Message {
	var messages = HistoryPrompt(contextWindow)
	messages = append(messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: compressPrompt,
	})

	return messages
}

func EstimateTokens(systemPrompt string, window conversation.Window, extraChars int) int {
	charnum := len(systemPrompt) + extraChars

	if window.Summary != nil {
		charnum += len(Summary(*window.Summary))
	}

	for _, h := range window.Events {
		charnum += len(Event(h))
	}

	token := charnum / 3
	return token
}
