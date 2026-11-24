package endpoint

import (
	"encoding/json"
	"yachiyo/yachiyo-shared/address"
)

type RemoteTool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Address     address.Address
}
