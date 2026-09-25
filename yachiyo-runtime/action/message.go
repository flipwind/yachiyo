package action

import (
	"yachiyo/yachiyo-runtime/address"
)

type Message struct {
	Content string
	Time    int64
	IsInitiative bool
	IsEmpty      bool

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *Message) action() {}