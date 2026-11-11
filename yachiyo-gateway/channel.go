package gateway

import (
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
)

type GatewayChannel struct {
	ToServer chan trigger.Trigger
	ToClient chan action.Action
}

func NewGatewayChannel() *GatewayChannel {
	return &GatewayChannel{
		ToServer: make(chan trigger.Trigger),
		ToClient: make(chan action.Action),
	}
}
