package core

import (
	"time"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		if c.LLMBusy == false {
			c.processTimetick()
		}
	}
}

func (c *Core) processTimetick() {
	// State press
	for _, s := range c.State.Drives() {
		s.Drive.Press((1.4 * 5) / (3600 * 1))
	}

	// Initiative
	timeNow := time.Now()
	var alonetime time.Duration

	if c.LastActiveTime == nil {
		alonetime = time.Since(time.Now())
	} else {
		alonetime = time.Since(*c.LastActiveTime)
	}
	daytime := timeNow.Sub(time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()))

	ylog.Debug("%s", c.InternalState())

	if c.Factors.Update(alonetime.Minutes(), daytime.Hours()) {
		ylog.Info("Initiative active.")

		// Relieve
		for _, s := range c.State.Drives() {
			s.Drive.Relieve(0.5)
			ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
		}

		c.Dispatch(&trigger.InitiativeMessage{})
	}
}
