package model

import "encoding/json"

type Envelope struct {
	Category string          `json:"category"`
	Data     json.RawMessage `json:"data"`
}
