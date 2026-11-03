package event

import (
	"fmt"
	"yachiyo/yachiyo-runtime/address"
)

type Message interface {
	Event
	message()
}

type UserMessageContent struct {
	BaseEvent
	Content  string
	Author   string
	Platform string
	Address  address.Address
}

type UserMessage struct {
	UserMessageContent
}

func (m UserMessage) message() {}
func (m UserMessage) history() {}
func (m UserMessageContent) Render() string {
	// e.g. <Iroha>(2030.07.21 16:16:08/Client) Good evening Yachiyo~
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Timestamp.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}

type AssistantMessageContent struct {
	BaseEvent
	Content []string

	Initiative bool
	Reply      bool
	Address    address.Address
}

type AssistantMessage struct {
	AssistantMessageContent
}

func (m AssistantMessage) message() {}
func (m AssistantMessage) history() {}
func (m AssistantMessageContent) Render() string {
	// e.g. <Message initiative: False><Reply: True>(2030.07.21 16:16:09) Iroha! Suki~
	msg := fmt.Sprintf("<Message initiative: %t><Reply: %t>(%s) ", m.Initiative, m.Reply, m.Timestamp.Format("2006.01.02 15:04:05"))
	if m.Reply {
		for _, message := range m.Content {
			msg += fmt.Sprintf("`%s` ", message)
		}
	} else {
		msg += fmt.Sprintf("<Reason: %s>", m.Content)
	}

	return msg
}
