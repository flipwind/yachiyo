package action

import (
	"yachiyo/yachiyo-runtime/address"
)

type Action interface {
	GetAddress() address.Address
	action()
}
