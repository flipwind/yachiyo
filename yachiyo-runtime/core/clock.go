package core

import (
	"time"
	"yachiyo/yachiyo-shared/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		c.processTimetick(1000 * time.Millisecond)
	}
}

func (c *Core) processTimetick(d time.Duration) {
	c.mu.Lock()

	if err := c.State.Advance(d, time.Now()); err != nil {
		ylog.Error("State advanced error: %v", err)
	}
	isTriggerInitiative := false // TODO: implement new initiative system

	if isTriggerInitiative {
		ylog.Info("Initiative active.")

		c.State.Relieve(d)
	}

	if err := c.checkpoint(); err != nil {
		ylog.Error("Saving checkpoint error: %v", err)
	}

	c.mu.Unlock()

	ylog.Debug("%s", c.InternalState())

	if isTriggerInitiative {
		c.Dispatch(&trigger.InitiativeMessage{})
	}
}
