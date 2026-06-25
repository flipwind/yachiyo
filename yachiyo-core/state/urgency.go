package state

import "strings"

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

func UrgencyIntroduce() string {
	return `Urgency is a way to measure the degree of something.
It consists of Trivial, Low, Medium, High, Urgent, Extreme.
These 6 types are from low to high.
If task asks generating Urgency, DON'T introduce other types.`
}

func (u *Urgency) FromString(t string) {
	t = strings.ToLower(t)
	switch t {
	case "trivial":
		*u = Trivial
	case "low":
		*u = Low
	case "medium":
		*u = Medium
	case "high":
		*u = High
	case "urgent":
		*u = Urgent
	case "extreme":
		*u = Extreme

	default:
		*u = Medium
	}
}

// State Change

func (u *Urgency) Increase() {
	if *u + 1 > Extreme {
		return
	}
	*u += 1
}

func (u *Urgency) Decrease() {
	if *u - 1 < Trivial {
		return
	}
	*u -= 1
}

func (u *Urgency) Relieve() {
	switch *u{
	case Trivial: 	*u = Trivial
	case Low:		*u = Low 
	case Medium:	*u = Low
	case High:		*u = Medium
	case Urgent:	*u = Medium
	case Extreme: 	*u = Low
	}
}

func (u *Urgency) Press(){
	switch *u{
	case Trivial:	*u = High
	case Low:		*u = Medium
	case Medium:	*u = Medium
	case High:		*u = Urgent
	case Urgent:	*u = Extreme
	case Extreme: 	*u = Extreme
	}
}
