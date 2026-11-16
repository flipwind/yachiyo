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
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := Snapshot{
		State: c.State.Snapshot(),
		Note:  c.conv.Note(),
	}

	return snap
}

// caller must hold c.mu
func (c *Core) checkpoint() error {
	return c.store.SaveCheckpoint(store.Checkpoint{
		SavedAt:  time.Now(),
		State:    c.State.Snapshot(),
		Note:     c.conv.Note(),
		Sequence: c.conv.Sequence(),
	})
}
