package trigger

import "yachiyo/yachiyo-shared/address"

type RuntimeStateRequest struct {
	Address address.Address
}

func (*RuntimeStateRequest) trigger() {}
