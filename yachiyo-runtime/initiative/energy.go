package initiative

import (
	"fmt"
	"yachiyo/yachiyo-util/ymath"
)

// An energy expression: x in [0, 24]; y in [0, 1]
//
// Value is the `y`;
// `x` is the time, unit `hour` in a day.
type Energy struct {
	Curve *ymath.ExprCurve
	Value float64
}

func NewEnergy(expression string) (Energy, error) {
	curve, err := ymath.New(expression, 0)
	if err != nil {
		return Energy{}, fmt.Errorf("Curve error: %w", err)
	}

	return Energy{
		Curve: curve,
		Value: 0,
	}, nil
}

func (e *Energy) Update(hour float64) error {
	value, err := e.Curve.Calculate(hour)
	if err != nil {
		return fmt.Errorf("Curve calculating error: %w", err)
	}

	e.Value = value
	return nil
}

func EnergyVerify(expression string) error {
	curve, err := ymath.New(expression, 0)

	if err != nil {
		return fmt.Errorf("Expression of energy is not vaild.")
	}

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
