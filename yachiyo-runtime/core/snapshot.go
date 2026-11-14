package core

import (
	"time"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/state"
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
		Note:  c.Note,
	}

	return snap
}

func (c *Core) checkpoint() error {
	return c.History.SaveCheckpoint(history.Checkpoint{
		Exist:       true,
		StateAt:     time.Now(),
		State:       c.State.Snapshot(),
		Note:        c.Note,
		NowSequence: c.History.NowSequence(),
	})
}
