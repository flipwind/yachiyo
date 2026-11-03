package trigger

import (
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/event"
)

type UserMessage struct {
	event.UserMessageContent
}

func (*UserMessage) trigger() {}

type MessageHistoryRequest struct {
	Address address.Address
}

func (*MessageHistoryRequest) trigger() {}
