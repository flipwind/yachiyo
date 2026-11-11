package action

import (
	"time"
)

type AssistantMessage struct {
	BaseAction
	Content    []string
	Timestamp  time.Time
	Initiative bool
	Reply      bool
}

func (*AssistantMessage) action() {}

// HistoryMessage is one entry of MessageHistory:
// either a UserHistoryMessage or a AssistantHistoryMessage.
type HistoryMessage interface {
	historyMessage()
}

type UserHistoryMessage struct {
	Content   string
	Timestamp time.Time
}

func (*UserHistoryMessage) historyMessage() {}

type AssistantHistoryMessage struct {
	Content    []string
	Timestamp  time.Time
	Initiative bool
	Reply      bool
}

func (*AssistantHistoryMessage) historyMessage() {}

type MessageHistory struct {
	BaseAction
	Messages []HistoryMessage
}

func (*MessageHistory) action() {}
