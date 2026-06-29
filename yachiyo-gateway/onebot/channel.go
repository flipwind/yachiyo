package onebot

import (
	"yachiyo/yachiyo-runtime/trigger"
)

type AdapterChannel struct {
	ToServer chan trigger.Message
	ToClient chan trigger.Message
}

func NewAdapterChannel() *AdapterChannel {
	return &AdapterChannel{
		ToServer: make(chan trigger.Message),
		ToClient: make(chan trigger.Message),
	}
}