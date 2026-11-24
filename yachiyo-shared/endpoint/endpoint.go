package endpoint

import "yachiyo/yachiyo-shared/address"

type Endpoint struct {
	Name       string
	ID         string
	Address    address.Address
	Capability Capability
	Tools      []RemoteTool
}
