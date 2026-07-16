package initiative

import (
	"fmt"
	"math/rand/v2"
)

type Factor struct {
	// TODO: curve f(x)
	Value  float64
	Max    float64
	Weight float64
}

func (f *Factor) percent() string {
	x := f.Value / f.Max
	if x > 1 {
		x = 1
	}

	if x < -1 {
		x = -1
	}

	return fmt.Sprintf("%.2f%%", x * 100)
}

func (f *Factor) score() float64 {
	x := f.Value / f.Max
	if x > 1 {
		x = 1
	}

	if x < -1 {
		x = -1
	}

	return x * f.Weight
}

type Factors struct {
	Threshold float64

	// Personailty
	Sociability Factor

	// Time
	AloneTime Factor

	// Environment
	// Daytime Factor

	// Bonus
	RandomBonus float64
}

func NewFactors(threshold float64, 
	sociability_value float64, sociability_max float64, sociability_weight float64,
	alonetime_value float64, alonetime_max float64, alonetime_weight float64) Factors {
	// TODO: Setting
	return Factors{
		Threshold: threshold,
		Sociability: Factor{
			Value:  sociability_value,
			Max:    sociability_max,
			Weight: sociability_weight,
		},
		AloneTime: Factor{
			Value:  alonetime_value,
			Max:    alonetime_max, // 2min
			Weight: alonetime_weight,
		},
	}
}

// Update the values, and return a probably initiative advice
//
// Note: alonetime should be a one-minute scale value, so as the daytime
func (fs *Factors) Update(alonetime float64, daytime float64) bool {
	fs.AloneTime.Value = alonetime
	// fs.Daytime.Value = daytime

	weights := fs.Sociability.Weight + fs.AloneTime.Weight 
	return fs.Score()/weights > fs.Threshold
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
		fs.RandomBonus
}

func (fs *Factors) String() string {
	all := fs.Sociability.score()+fs.AloneTime.score()+fs.RandomBonus
	weights := fs.Sociability.Weight + fs.AloneTime.Weight

	
	return fmt.Sprintf(`Sociability: %v,
AloneTime: %v,
Daytime: not available,
RandomBonus: %v,

factors sum = %v
accordingly initiative possibility = %.2f%%`, fs.Sociability.percent(), fs.AloneTime.percent(), fs.RandomBonus,
		all, all/weights*100)
}
