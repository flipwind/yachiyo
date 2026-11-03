package action

import (
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/event"
)

type AssistantMessage struct {
	event.AssistantMessageContent
}

func (m *AssistantMessage) action() {}
func (m *AssistantMessage) GetAddress() address.Address {
	return m.Address
}

type MessageHistory struct {
	Messages []event.Message
	Address  address.Address
}

func (m *MessageHistory) action() {}
func (m *MessageHistory) GetAddress() address.Address {
	return m.Address
}
