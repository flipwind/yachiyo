package action

import (
	"yachiyo/yachiyo-runtime/address"
)

type Message interface {
	Message()
}

type UserMessage struct {
	Content string
	Time    int64

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *UserMessage) action() {}
func (*UserMessage) Message(){}

type AssistantMessage struct {
	Content string
	Time    int64
	Initiative bool
	Reply      bool

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *AssistantMessage) action() {}
func (*AssistantMessage) Message(){}

type MessageHistory struct {
	Messages []Message
	Address address.Address
}

func (*MessageHistory) action(){}