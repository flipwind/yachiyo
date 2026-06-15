package chat

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func ContentLoggerFormat(m *Message) string {
	return fmt.Sprintf("<%v>[%v] %v", m.Role, time.Unix(m.Timestamp, 0).Format("2006.01.02 15:04:05"), m.Content)
}

func BuildMessage(session_id string, role string, content string, timestamp int64) *Message {
	return &Message{
		Uuid: uuid.NewString(),
		SessionId: session_id,
		Role: role,
		Content: content,
		Timestamp: timestamp,
	}
}