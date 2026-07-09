package core

import (
	"time"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Clock() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for tick := range ticker.C {
		c.Pipe.Raw <- &trigger.TimeTick{
			Time: tick,
		}
	}
}