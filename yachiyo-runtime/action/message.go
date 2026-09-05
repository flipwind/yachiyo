package action

import "yachiyo/yachiyo-runtime/address"

type Message struct {
	Empty   bool
	Content string
	Time    int64

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *Message) action() {}