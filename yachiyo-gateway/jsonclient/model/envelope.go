package model

import "encoding/json"

type Envelope struct {
	Category string          `json:"category"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
}

type DataPack interface {
	DataPack()
}
