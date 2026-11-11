package prompt

import (
	"fmt"
	"yachiyo/yachiyo-shared/trigger"
)

func RenderUserMessage(m trigger.UserMessage) string {
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Timestamp.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}
