package action

import "yachiyo/yachiyo-runtime/address"

type RuntimeState struct {
	Content string

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *RuntimeState) action() {}