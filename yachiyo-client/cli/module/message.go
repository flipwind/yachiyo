package module

import (
	"fmt"
	"time"
)

type Message struct {
	time    time.Time
	sender  string
	content string
}

func (m *Message) String() string {
	return fmt.Sprintf("[%s] @%s > %s", m.time.Format("15:04:05"), m.sender, m.content)
}