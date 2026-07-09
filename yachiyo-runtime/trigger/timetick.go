package trigger

import "time"

type TimeTick struct {
	Time time.Time
}

func (t *TimeTick) trigger() {}
