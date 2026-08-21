package initiative

import (
	"fmt"
	"math/rand/v2"
	"yachiyo/yachiyo-runtime/config"
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

func newFactor(f config.FactorConfig) Factor {
	// TODO: error

	curve, err := ymath.New(*f.Curve, *f.DefaultValue)
	if err != nil {
		ylog.Error("Curve error: %v", err)
	}

	return Factor{
		Curve: *curve,
		Value: *f.DefaultValue,
		Max: *f.Max,
		Weight: *f.Weight,
	}
}


func NewFactors(threshold float64,
	sociability config.FactorConfig,
	alonetime config.FactorConfig,
	daytime config.FactorConfig) Factors {
	// TODO: Setting
	return Factors{
		Threshold: threshold,
		Sociability: newFactor(sociability),
		AloneTime: newFactor(alonetime),
		Daytime: newFactor(daytime),
	}
}

// Update the values, and return a probably initiative advice
//
// Note: alonetime should be a one-minute scale value, so as the daytime
func (fs *Factors) Update(alonetime float64, daytime float64) {
	fs.AloneTime.Value = alonetime
	fs.Daytime.Value = daytime
}

func (fs *Factors) InitiativeAdvice() bool {
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
