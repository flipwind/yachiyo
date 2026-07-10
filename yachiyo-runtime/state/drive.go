package state

import "math"

// Why there are so many MAGIC NUMBERS...
// I *hate* these MAGIC NUMBERS.
// 
// And,
// We may need a better curve.
//
// TODO: a better curve

type Drive struct {
	Value float64	// the `x` on the curve, not `y` value
}

func curveF(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func (d *Drive) Urgency() Urgency {
	y := curveF(d.Value)
	switch {
	case y < 0.15:
		return Trivial
	case y < 0.35:
		return Low
	case y < 0.65:
		return Medium
	case y < 0.8:
		return High
	case y < 0.95:
		return Urgent
	default:
		return Extreme
	}
}

func (d *Drive) String() string {
	u := d.Urgency()
	return u.String()
}

func UrgencyValue(u Urgency) float64 {
	var value float64
	switch u {
	case Trivial:
		value = -2
	case Low:
		value = -0.8
	case Medium:
		value = 0
	case High:
		value = 0.85
	case Urgent:
		value = 2.2
	case Extreme:
		value = 3
	}
	return value
}

func UrgencyDrive(u Urgency) Drive {
	value := UrgencyValue(u)
	return Drive{
		Value: value,
	}
}

// State change

func (d *Drive) Increase() {
	d.Value += 0.2
}

func (d *Drive) Decrease() {
	d.Value -= 0.2
}

func (d *Drive) State(u Urgency) {
	value := UrgencyValue(u)
	d.Value = value
}

// We advise speedK should be in 0.0 ~ 1.0

func (d *Drive) Press(speedK float64) {
	d.Value += 0.2 * speedK
}

func (d *Drive) Relieve(speedK float64) {
	d.Value -= 0.2 * speedK
}