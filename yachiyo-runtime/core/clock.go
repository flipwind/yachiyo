package core

import (
	"time"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		c.processTimetick()
	}
}

func (c *Core) processTimetick() {
	c.mu.Lock()

	// State press
	for _, s := range c.State.Drives() {
		s.Drive.Press((1.4 * 5) / (3600 * 1))
	}

	// Affect Engine
	c.Affect.Simulate(1)

	// Initiative
	timeNow := time.Now()

	c.Energy.Update(float64(float64(timeNow.Hour()) + float64(timeNow.Minute())/60.0))
	isTriggerInitiative := false // TODO: implement new initiative system

	if isTriggerInitiative {
		ylog.Info("Initiative active.")

		// Relieve
		for _, s := range c.State.Drives() {
			s.Drive.Relieve(0.5)
			ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
		}
	}

	c.mu.Unlock()

	ylog.Debug("%s", c.InternalState())

	if isTriggerInitiative {
		c.Dispatch(&trigger.InitiativeMessage{})
	}
}
