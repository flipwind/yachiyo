package trigger

import "yachiyo/yachiyo-runtime/address"

type RuntimeStateRequest struct {
	Address address.Address
}

func (*RuntimeStateRequest) trigger() {}