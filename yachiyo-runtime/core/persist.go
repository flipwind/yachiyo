package core

import (
	"fmt"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
)

func (c *Core) newMeta(snap state.Snapshot, t time.Time) event.Meta {
	return event.Meta{
		ID:        uuid.NewV7(),
		Sequence:  c.conv.NextSequence(),
		Timestamp: t,
		Snapshot:  snap,
	}
}

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

func (c *Core) saveCheckpoint() {
	if err := c.checkpoint(); err != nil {
		ylog.Error("Checkpoint saving error: %v", err)
	}
}

func (c *Core) record(e event.Event) error {
	if err := c.store.Append(e); err != nil {
		return fmt.Errorf("storage saving event error: %w", err)
	}

	c.conv.Append(e)
	return nil
}

func (c *Core) commitSummary(summary conversation.Summary) error {
	if err := c.store.SaveSummary(summary); err != nil {
		return fmt.Errorf("storage saving summary error: %w", err)
	}

	c.conv.SetSummary(summary)
	return nil
}

func (c *Core) internalState() string {
	msg := fmt.Sprintf("Received timetick %s\n", time.Now().Format("15:04:05"))

	snap := c.snapshot().State
	for _, s := range snap.Drives.All() {
		msg += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	msg += fmt.Sprintf("Energy: %v\n", snap.Energy)

	return msg
}
