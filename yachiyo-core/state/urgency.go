package state

type Urgency int

const (
	Trivial Urgency = iota
	Low
	Medium
	High
	Urgent
	Extreme
)

func (u *Urgency) String() string {
	switch *u {
	case Trivial:
		return "Trivial"
	case Low:
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	case Urgent:
		return "Urgent"
	case Extreme:
		return "Extreme"
	}
	return "Unknown"
}
