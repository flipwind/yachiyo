package event

import (
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/state"
)

type Event interface {
	GetSequence() int64
	GetID() uuid.UUID
	Render() string
}

type Snapshot struct {
	Timestamp      time.Time
	Affect         state.Engine
	Drives         state.Drives
	Energy         float64
	LastActiveTime time.Time
}

type BaseEvent struct {
	ID        uuid.UUID
	Sequence  int64
	Timestamp time.Time
	Snapshot  Snapshot
}

func (b BaseEvent) GetSequence() int64 {
	return b.Sequence
}

func (b BaseEvent) GetID() uuid.UUID {
	return b.ID
}
