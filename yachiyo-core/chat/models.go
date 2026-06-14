package chat

import (
	"fmt"
	"time"
)

func (m *Message) ContentBuild() {
	content := fmt.Sprintf("<%v>[%v] %v", m.Role, time.Unix(m.Timestamp, 0).Format("2006.01.02 15:04:05"), m.Content)
	m.Content = content
}