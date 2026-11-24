package gateway

import (
	"context"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/endpoint"
	"yachiyo/yachiyo-shared/trigger"
)

type Service interface {
	SchemeName() string
	Listen(port int64, submit func(trig trigger.Trigger))
	Send(a action.Action) error
}

type EndpointProvider interface {
	Clients() []endpoint.Endpoint
}

type ToolExecutor interface {
	UseTool(ctx context.Context, addr address.Address, reqID, toolName, payload string) (string, error)
}
