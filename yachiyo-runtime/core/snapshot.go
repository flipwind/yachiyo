package core

import (
	"time"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
)

// Core snapshot

type Snapshot struct {
	Drives         state.Drives
	Affect         state.Engine
	Determination  state.Determination
	Energy         initiative.Energy
	Note           string
	LastActiveTime time.Time
}

func (c *Core) snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := Snapshot{
		Drives:         c.Drives,
		Affect:         c.Affect,
		Determination:  c.Determination,
		Energy:         c.Energy,
		Note:           c.Note,
		LastActiveTime: c.LastActiveTime,
	}

	return snap
}

func (c *Core) checkpoint() error {
	return c.History.SaveCheckpoint(history.Checkpoint{
		Exist:          true,
		StateAt:        time.Now(),
		Drives:         c.Drives,
		Determination:  c.Determination,
		Energy:         c.Energy,
		Affect:         c.Affect,
		Note:           c.Note,
		LastActiveTime: c.LastActiveTime,
		NowSequence:    c.History.NowSequence(),
	})
}

func (snap Snapshot) toHistorySnapshot(time time.Time) event.Snapshot {
	return event.Snapshot{
		Timestamp:      time,
		Affect:         snap.Affect,
		Drives:         snap.Drives,
		Energy:         snap.Energy.Value,
		LastActiveTime: snap.LastActiveTime,
	}
}
