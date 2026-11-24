package hub

import (
	"context"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/endpoint"
)

type Hub interface {
	Deliver(a action.Action)
	AvailableAddresses(cap string) ([]address.Address, error)
	ClientTools() []endpoint.RemoteTool
	SendMessage(a action.AssistantMessage) error
	UseTool(ctx context.Context, addr address.Address, reqID, toolName, payload string) (string, error)
}
