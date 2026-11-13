package state

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-util/ymath"
)

// An energy expression: x in [0, 24]; y in [0, 1]
//
// Value is the `y`;
// `x` is the time, unit `hour` in a day.
type Energy struct {
	curve *ymath.ExprCurve
	value float64
}

func NewEnergy(expression string) (Energy, error) {
	curve, err := ymath.New(expression, 0)
	if err != nil {
		return Energy{}, fmt.Errorf("Curve error: %w", err)
	}

	if err := verify(curve); err != nil {
		return Energy{}, fmt.Errorf("Curve error: %w", err)
	}

	return Energy{
		curve: curve,
		value: 0,
	}, nil
}

func (e *Energy) Update(t time.Time) error {
	hour := float64(t.Hour()) + float64(t.Minute())/60.0

	value, err := e.curve.Calculate(hour)
	if err != nil {
		return fmt.Errorf("Curve calculating error: %w", err)
	}

	e.value = value
	return nil
}

func (e Energy) Value() float64 {
	return e.value
}

func verify(curve *ymath.ExprCurve) error {
	for i := 0.0; i <= 24; i += 0.1 {
		value, err := curve.Calculate(i)
		if err != nil {
			return fmt.Errorf("Energy calculated error: %w", err)
		}
		if !(value >= 0 && value <= 1) {
			return fmt.Errorf("Energy not vaild, should be in [0, 1], but was %v when { x = %v }", value, i)
		}
	}

	return nil
}
