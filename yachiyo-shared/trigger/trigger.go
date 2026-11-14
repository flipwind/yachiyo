package trigger

import (
	"time"
	"yachiyo/yachiyo-shared/address"
)

type Trigger interface {
	trigger()
}

type UserMessage struct {
	Author    string
	Platform  string
	Content   string
	Timestamp time.Time
	Address   address.Address
}

func (*UserMessage) trigger() {}

// TODO: taskify
type InitiativeMessage struct{}

func (*InitiativeMessage) trigger() {}

type MessageHistoryRequest struct {
	Address address.Address
}

func (*MessageHistoryRequest) trigger() {}

type RuntimeStateRequest struct {
	Address address.Address
}

func (*RuntimeStateRequest) trigger() {}
