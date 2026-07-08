package gateway

import (
	"yachiyo/yachiyo-runtime/trigger"
)

type GatewayChannel struct {
	ToServer chan trigger.Trigger
	ToClient chan trigger.Trigger
}

func NewGatewayChannel() *GatewayChannel {
	return &GatewayChannel{
		ToServer: make(chan trigger.Trigger),
		ToClient: make(chan trigger.Trigger),
	}
}