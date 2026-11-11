package trigger

import (
	"fmt"
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

func (UserMessage) trigger() {}

func (m UserMessage) Render() string {
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Timestamp.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}

// TODO: taskify
type InitiativeMessage struct {}

func (InitiativeMessage) trigger() {}

type MessageHistoryRequest struct {
	Address address.Address
}

func (MessageHistoryRequest) trigger() {}

type RuntimeStateRequest struct {
	Address address.Address
}

func (RuntimeStateRequest) trigger() {}
