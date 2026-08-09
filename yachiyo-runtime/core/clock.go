package core

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for tick := range ticker.C {
		if c.LLMBusy == false {
			c.processTimetick(&trigger.TimeTick{
				Time: tick,
			})
		}
	}
}

func (c *Core) processTimetick(t *trigger.TimeTick) {
	// TODO: too ugly, need simplify

	DebugMessage := fmt.Sprintf("Received timetick %s\n", t.Time.Format("15:04:05"))

	// State press
	for _, s := range c.State.Drives() {
		s.Drive.Press((1.4 * 5) / (3600 * 1))
		DebugMessage += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
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
	DebugMessage += fmt.Sprintf("factors: %v\n", c.Factors.String())

	if c.Factors.Update(alonetime.Minutes(), daytime.Hours()) {
		ylog.Info("Initiative active.")

		// Relieve
		for _, s := range c.State.Drives() {
			s.Drive.Relieve(0.5)
			ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
		}

		c.Dispatch(&trigger.InitiativeMessage{})
	}

	ylog.Debug("%s", DebugMessage)
	c.Distribution(&action.Status{
		Content: DebugMessage,
	})
}
