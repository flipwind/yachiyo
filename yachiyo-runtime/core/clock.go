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

	// State press
	for _, s := range c.Drives.All() {
		s.Drive.Press((1.4 * 5) / (3600 * 1) * d.Seconds())
	}

	// Affect Engine
	c.Affect.Simulate(d.Seconds())

	// Initiative
	if err := c.Energy.Update(time.Now()); err != nil {
		ylog.Error("Energy updated error: %v", err)
	}
	isTriggerInitiative := false // TODO: implement new initiative system

	if isTriggerInitiative {
		ylog.Info("Initiative active.")

		// Relieve
		for _, s := range c.Drives.All() {
			s.Drive.Relieve(0.5 * d.Seconds())
			ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
		}
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
