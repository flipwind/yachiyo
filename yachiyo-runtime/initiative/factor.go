package initiative

import (
	"fmt"
	"math/rand/v2"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/ymath"
)

var ylog = logger.New("Yachiyo.Initiative")

type Factor struct {
	Curve  ymath.ExprCurve
	Value  float64
	Max    float64
	Weight float64
}

func (f *Factor) percent() string {
	x := f.score() / f.Weight / f.Max

	return fmt.Sprintf("%.2f%%", x*100)
}

func (f *Factor) score() float64 {
	y, err := f.Curve.Calculate(f.Value)

	if err != nil {
		ylog.Error("Calculate curve <%v> error: %v", f.Curve.Expression, err)
	}

	return y * f.Weight
}

type Factors struct {
	Threshold float64

	// Personailty
	Sociability Factor

	// Time
	AloneTime Factor

	// Environment
	Daytime Factor

	// Bonus
	RandomBonus float64
}

func NewFactors(threshold float64,
	sociability_curve string, sociability_value float64, sociability_max float64, sociability_weight float64,
	alonetime_curve string, alonetime_value float64, alonetime_max float64, alonetime_weight float64,
	daytime_curve string, daytime_value float64, daytime_max float64, daytime_weight float64) Factors {
	// TODO: Setting
	// TODO: adaptive

	// curves
	// TODO: error
	curve_sociability, err := ymath.New(sociability_curve, sociability_value)
	if err != nil {
		ylog.Error("Curve error: %v", err)
	}

	curve_alonetime, err := ymath.New(alonetime_curve, alonetime_value)
	if err != nil {
		ylog.Error("Curve error: %v", err)
	}

	curve_daytime, err := ymath.New(daytime_curve, daytime_value)
	if err != nil {
		ylog.Error("Curve error: %v", err)
	}

	return Factors{
		Threshold: threshold,
		Sociability: Factor{
			Curve:  *curve_sociability,
			Value:  sociability_value,
			Max:    sociability_max,
			Weight: sociability_weight,
		},
		AloneTime: Factor{
			Curve:  *curve_alonetime,
			Value:  alonetime_value,
			Max:    alonetime_max,
			Weight: alonetime_weight,
		},
		Daytime: Factor{
			Curve:  *curve_daytime,
			Value:  daytime_value,
			Max:    daytime_max,
			Weight: daytime_weight,
		},
	}
}

// Update the values, and return a probably initiative advice
//
// Note: alonetime should be a one-minute scale value, so as the daytime
func (fs *Factors) Update(alonetime float64, daytime float64) bool {
	fs.AloneTime.Value = alonetime
	fs.Daytime.Value = daytime

	max := fs.Sociability.Max*fs.Sociability.Weight +
		fs.AloneTime.Max*fs.AloneTime.Weight +
		fs.Daytime.Max*fs.Daytime.Weight
	return fs.Score()/max > fs.Threshold
}

func randomBonus() float64 {
	// randomBonus is in (-0.05, 0.05)
	signT := rand.IntN(2) == 1
	var sign float64
	if signT {
		sign = 1
	} else {
		sign = -1
	}
	return rand.Float64() / 20 * sign
}

func (fs *Factors) Score() float64 {
	fs.RandomBonus = randomBonus()

	return fs.Sociability.score() +
		fs.AloneTime.score() +
		fs.Daytime.score() +
		fs.RandomBonus
}

func (fs *Factors) String() string {
	all := fs.Sociability.score() + fs.AloneTime.score() + fs.Daytime.score() + fs.RandomBonus
	max := fs.Sociability.Max*fs.Sociability.Weight +
		fs.AloneTime.Max*fs.AloneTime.Weight +
		fs.Daytime.Max*fs.Daytime.Weight

	return fmt.Sprintf(`Sociability: %v,
AloneTime: %v,
Daytime: %v,
RandomBonus: %v,

factors sum = %v
accordingly initiative possibility = %.2f%%`, fs.Sociability.percent(), fs.AloneTime.percent(), fs.Daytime.percent(), fs.RandomBonus,
		all, all/max*100)
}
