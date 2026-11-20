package core

import (
	"time"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
)

// Core snapshot

type Snapshot struct {
	State state.Snapshot
	Note  string
}

func (c *Core) snapshot() Snapshot {
	snap := Snapshot{
		State: c.state.Snapshot(),
		Note:  c.conv.Note(),
	}

	return snap
}

func (c *Core) checkpoint() error {
	return c.store.SaveCheckpoint(store.Checkpoint{
		SavedAt:  time.Now(),
		State:    c.state.Snapshot(),
		Note:     c.conv.Note(),
		Sequence: c.conv.Sequence(),
	})
}
