package action

import (
	"yachiyo/yachiyo-shared/address"
)

type Action interface {
	GetAddress() address.Address
	action()
}
