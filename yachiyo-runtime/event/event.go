package event

import (
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/address"
)

type Event interface {
	event()
	GetMeta() Meta
}

type Meta struct {
	ID        uuid.UUID
	Sequence  int64
	Timestamp time.Time
	Snapshot  state.Snapshot
}

func (m Meta) GetMeta() Meta { return m }

type UserMessage struct {
	Meta
	Content  string
	Author   string
	Platform string
	Address  address.Address
}

func (UserMessage) event() {}

type AssistantMessage struct {
	Meta
	Content    []string
	Initiative bool
	Reply      bool
	Address    address.Address
}

func (AssistantMessage) event() {}
