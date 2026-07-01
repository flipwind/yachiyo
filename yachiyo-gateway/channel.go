package gateway

import (
	"yachiyo/yachiyo-runtime/trigger"
)

type GatewayChannel struct {
	ToServer chan trigger.Message
	ToClient chan trigger.Message
}

func NewGatewayChannel() *GatewayChannel {
	return &GatewayChannel{
		ToServer: make(chan trigger.Message),
		ToClient: make(chan trigger.Message),
	}
}