package trigger

import (
	"fmt"
	"time"
)

type Message struct {
	Type     string
	Author   string
	Platform string
	Content  string
	Time     int64

	Address Address // The address of the message. e.g. onebot://group/12345
}

func (m *Message) String() string {
	return fmt.Sprintf("<[%s] %s>(%s/%s) %s",
		m.Type,
		m.Author,
		time.Unix(m.Time, 0).Format("2006.01.02 15:04:05"),
		m.Platform,
		m.Content,
	)
}