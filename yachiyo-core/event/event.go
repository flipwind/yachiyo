package event

import (
	"fmt"
	"time"
)

type Event interface {
	event()
}

type Message struct {
	Type 	string
	Author  string
	Source  string
	Content string
	Time    int64
	Payload any
}

func (m *Message) String() string {
	return fmt.Sprintf("<[%s] %s>(%s/%s) %s",
		m.Type,
		m.Author,
		time.Unix(m.Time, 0).Format("2006.01.02 15:04:05"),
		m.Source,
		m.Content,
	)
}

type UserMessageEvent struct {
	Message Message
}

func (*UserMessageEvent) event() {}
