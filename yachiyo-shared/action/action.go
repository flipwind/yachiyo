package action

import (
	"yachiyo/yachiyo-shared/address"
)

// Action is what the runtime sends to a gateway.
type Action interface {
	action()
	GetAddress() address.Address
}

type BaseAction struct {
	Address address.Address
}

func (a *BaseAction) GetAddress() address.Address {
	return a.Address
}
