package action

import "yachiyo/yachiyo-runtime/address"

type Action interface {
	action()
	GetAddress() address.Address
}
