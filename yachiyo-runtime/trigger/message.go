package trigger

import (
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-shared/address"
)

type UserMessage struct {
	event.UserMessageContent
}

func (*UserMessage) trigger() {}

type MessageHistoryRequest struct {
	Address address.Address
}

func (*MessageHistoryRequest) trigger() {}
