package core

import (
	"time"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {

		c.mu.Lock()
		busy := c.LLMBusy
		c.mu.Unlock()
		if !busy {
			c.processTimetick()
		}
	}
}

func (c *Core) processTimetick() {
	c.mu.Lock()

	// State press
	for _, s := range c.State.Drives() {
		s.Drive.Press((1.4 * 5) / (3600 * 1))
	}

	// Initiative
	timeNow := time.Now()

	alonetime := time.Since(c.LastActiveTime)
	daytime := timeNow.Sub(time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()))

	c.Factors.Update(alonetime.Minutes(), daytime.Hours())
	isTriggerInitiative := c.Factors.InitiativeAdvice()

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
