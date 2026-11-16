package store

import (
	"time"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/state"
)

type Checkpoint struct {
	SavedAt  time.Time
	State    state.Snapshot
	Note     string
	Sequence int64
}

type Restored struct {
	Events     []event.Event
	Checkpoint *Checkpoint
	Summary    *conversation.Summary
}

type Store interface {
	Load() (Restored, error)
	Append(e event.Event) error
	SaveCheckpoint(c Checkpoint) error
	SaveSummary(s conversation.Summary) error
	Close() error
}
