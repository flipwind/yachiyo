// The implement of this file is generally based on ALMA.
// See https://doi.org/10.1145/1082473.1082478, ALMA – A Layered Model of Affect
package state

import (
	"fmt"
	"math"
	"yachiyo/yachiyo-util/ymath"
)

// The PAD Space.
//
// X = Pleasure, Y = Arousal, Z = Dominance
type PAD struct {
	ymath.Vector3
}

func (p PAD) Pleasure() float64  { return p.X }
func (p PAD) Arousal() float64   { return p.Y }
func (p PAD) Dominance() float64 { return p.Z }

// Return a readable string of PAD, using types in the papper.
func (p PAD) Category() string {
	var strength string

	length := p.Length()
	sqrt3 := math.Sqrt(3)

	// simple implement, sqrt(3) may not be suitable
	if length < sqrt3/3 {
		strength = "slightly"
	} else if length < sqrt3/3*2 {
		strength = "moderately"
	} else {
		strength = "highly"
	}

	return fmt.Sprintf("<%s %s>", strength, p.Octants())
}

func (p PAD) Octants() string {
	var types string
	switch {
	case p.X > 0 && p.Y > 0 && p.Z > 0:
		types = "Exuberant"
	case p.X > 0 && p.Y > 0 && p.Z < 0:
		types = "Dependent"
	case p.X > 0 && p.Y < 0 && p.Z > 0:
		types = "Relaxed"
	case p.X > 0 && p.Y < 0 && p.Z < 0:
		types = "Docile"
	case p.X < 0 && p.Y < 0 && p.Z < 0:
		types = "Bored"
	case p.X < 0 && p.Y < 0 && p.Z > 0:
		types = "Disdainful"
	case p.X < 0 && p.Y > 0 && p.Z < 0:
		types = "Anxious"
	case p.X < 0 && p.Y > 0 && p.Z > 0:
		types = "Hostile"
	default:
		types = "Undefined"
	}
	return types
}

func (p PAD) String() string {
	measure := func(num float64) string {
		num = math.Abs(num)
		if num == 0 {
			return "neutral"
		} else if num < 1.0/3 {
			return "slightly"
		} else if num < 2.0/3 {
			return "moderately"
		} else {
			return "highly"
		}
	}

	var _p, _a, _d string
	if p.X > 0 {
		_p = "pleasant"
	} else if p.X < 0 {
		_p = "unpleasant"
	} else {
		_p = "pleasure"
	}

	if p.Y > 0 {
		_a = "aroused"
	} else if p.Y < 0 {
		_a = "unaroused"
	} else {
		_a = "arousal"
	}

	if p.Z > 0 {
		_d = "dominant"
	} else if p.Z < 0 {
		_d = "submissive"
	} else {
		_d = "dominance"
	}

	return fmt.Sprintf("<%s: %s(%f), %s: %s(%f), %s: %s(%f)>",
		_p, measure(p.X), p.X,
		_a, measure(p.Y), p.Y,
		_d, measure(p.Z), p.Z,
	)
}

// OCC emotions in PAD space, using types in the papper.
type OCCEmotionType struct {
	ymath.Vector3
	Name string
}

type OCCEmotion struct {
	Type      OCCEmotionType
	Intensity float64 // should be in [0, 1]
}

var (
	Admiration     = OCCEmotionType{ymath.Vector3{X: 0.5, Y: 0.3, Z: -0.2}, "Admiration"}
	Anger          = OCCEmotionType{ymath.Vector3{X: -0.51, Y: 0.59, Z: 0.25}, "Anger"}
	Disliking      = OCCEmotionType{ymath.Vector3{X: -0.4, Y: 0.2, Z: 0.1}, "Disliking"}
	Disappointment = OCCEmotionType{ymath.Vector3{X: -0.3, Y: 0.1, Z: -0.4}, "Disappointment"}
	Distress       = OCCEmotionType{ymath.Vector3{X: -0.4, Y: -0.2, Z: -0.5}, "Distress"}
	Fear           = OCCEmotionType{ymath.Vector3{X: -0.64, Y: 0.60, Z: -0.43}, "Fear"}
	FearsConfirmed = OCCEmotionType{ymath.Vector3{X: -0.5, Y: -0.3, Z: -0.7}, "FearsConfirmed"}
	Gloating       = OCCEmotionType{ymath.Vector3{X: 0.3, Y: -0.3, Z: -0.1}, "Gloating"}
	Gratification  = OCCEmotionType{ymath.Vector3{X: 0.6, Y: 0.5, Z: 0.4}, "Gratification"}
	Gratitude      = OCCEmotionType{ymath.Vector3{X: 0.4, Y: 0.2, Z: -0.3}, "Gratitude"}
	HappyFor       = OCCEmotionType{ymath.Vector3{X: 0.4, Y: 0.2, Z: 0.2}, "HappyFor"}
	Hate           = OCCEmotionType{ymath.Vector3{X: -0.6, Y: 0.6, Z: 0.3}, "Hate"}
	Hope           = OCCEmotionType{ymath.Vector3{X: 0.2, Y: 0.2, Z: -0.1}, "Hope"}
	Joy            = OCCEmotionType{ymath.Vector3{X: 0.4, Y: 0.2, Z: 0.1}, "Joy"}
	Liking         = OCCEmotionType{ymath.Vector3{X: 0.40, Y: 0.16, Z: -0.24}, "Liking"}
	Love           = OCCEmotionType{ymath.Vector3{X: 0.3, Y: 0.1, Z: 0.2}, "Love"}
	Pity           = OCCEmotionType{ymath.Vector3{X: -0.4, Y: -0.2, Z: -0.5}, "Pity"}
	Pride          = OCCEmotionType{ymath.Vector3{X: 0.4, Y: 0.3, Z: 0.3}, "Pride"}
	Relief         = OCCEmotionType{ymath.Vector3{X: 0.2, Y: -0.3, Z: 0.4}, "Relief"}
	Remorse        = OCCEmotionType{ymath.Vector3{X: -0.3, Y: 0.1, Z: -0.6}, "Remorse"}
	Reproach       = OCCEmotionType{ymath.Vector3{X: -0.3, Y: -0.1, Z: 0.4}, "Reproach"}
	Resentment     = OCCEmotionType{ymath.Vector3{X: -0.2, Y: -0.3, Z: -0.2}, "Resentment"}
	Satisfaction   = OCCEmotionType{ymath.Vector3{X: 0.3, Y: -0.2, Z: 0.4}, "Satisfaction"}
	Shame          = OCCEmotionType{ymath.Vector3{X: -0.3, Y: 0.1, Z: -0.6}, "Shame"}
)

var OCCEmotionMap = map[string]OCCEmotionType{
	"Admiration":     Admiration,
	"Anger":          Anger,
	"Disliking":      Disliking,
	"Disappointment": Disappointment,
	"Distress":       Distress,
	"Fear":           Fear,
	"FearsConfirmed": FearsConfirmed,
	"Gloating":       Gloating,
	"Gratification":  Gratification,
	"Gratitude":      Gratitude,
	"HappyFor":       HappyFor,
	"Hate":           Hate,
	"Hope":           Hope,
	"Joy":            Joy,
	"Liking":         Liking,
	"Love":           Love,
	"Pity":           Pity,
	"Pride":          Pride,
	"Relief":         Relief,
	"Remorse":        Remorse,
	"Reproach":       Reproach,
	"Resentment":     Resentment,
	"Satisfaction":   Satisfaction,
	"Shame":          Shame,
}

// The Big Five personality traits.
// Every value should be in [0, 1]
//
// https://en.wikipedia.org/wiki/Big_Five_personality_traits
type BigFive struct {
	Openness          float64
	Conscientiousness float64
	Extraversion      float64
	Agreeableness     float64
	Neuroticism       float64
}

// Describe a point floating in the space.
type Point struct {
	Position PAD // (x, y, z)
}

func (p Point) Octants() string    { return p.Position.Octants() }
func (p Point) Category() string   { return p.Position.Category() }
func (p Point) Pleasure() float64  { return p.Position.Pleasure() }
func (p Point) Arousal() float64   { return p.Position.Arousal() }
func (p Point) Dominance() float64 { return p.Position.Dominance() }

func (p *Point) Move(step ymath.Vector3) {
	result := p.Position.Add(step)
	p.Position = PAD{Vector3: ymath.Vector3{
		X: Clamp(result.X, -1, 1),
		Y: Clamp(result.Y, -1, 1),
		Z: Clamp(result.Z, -1, 1),
	}}
}

func Clamp(n, a, b float64) float64 {
	if a > b {
		a, b = b, a
	}
	if n < a {
		return a
	}
	if n > b {
		return b
	}
	return n
}

type Engine struct {
	Default  PAD          // The default mood
	Current  Point        // The current point
	Emotions []OCCEmotion // Current OCCEmotion
}

func (e Engine) String() string {
	moods := ""
	for _, mood := range e.Emotions {
		moods += fmt.Sprintf("	- %s: %f\n", mood.Type.Name, mood.Intensity)
	}

	if moods == "" {
		moods = "	- Normal"
	}

	return fmt.Sprintf(`Mood Octants: %s
PAD Space: %s
Active Moods:
%s`, e.Current.Position.Category(), e.Current.Position.String(), moods)
}

// Recommend to use only when BigFive values significantly changed.
func NewEngine(b BigFive) Engine {
	defaultMood := ymath.Vector3{
		X: 0.21*b.Extraversion + 0.59*b.Agreeableness + 0.19*b.Neuroticism,
		Y: 0.15*b.Openness + 0.30*b.Agreeableness - 0.57*b.Neuroticism,
		Z: 0.25*b.Openness + 0.17*b.Conscientiousness + 0.60*b.Extraversion - 0.32*b.Agreeableness,
	}

	return Engine{
		Default: PAD{
			Vector3: defaultMood,
		},
		Current: Point{
			Position: PAD{defaultMood},
		},
	}
}

func (e *Engine) VirtualMoodCenter() ymath.Vector3 {
	var result ymath.Vector3
	for _, emotion := range e.Emotions {
		vector := emotion.Type.Scale(emotion.Intensity)
		result = result.Add(vector)
	}
	return result
}

func (e *Engine) Press(emotion OCCEmotion) {
	e.Emotions = append(e.Emotions, emotion)
}

// Simulate, and step according seconds
func (e *Engine) Simulate(seconds float64) {
	// 1. Decay

	const halflife float64 = 0.55 * 60 // halflife time, use seconds as unit
	const boundary float64 = 2e-2
	// On these values, a emotion, which origin intensity is 1,
	// disappeared in about 3 minutes.

	const movespeed float64 = 1.0 / (60 * 10)                  // move from a quadrant to another, spending 10 minutes
	const returnspeed float64 = 1.7320508075688772 / (60 * 20) // move from space boundary (1, 1, 1) to (0, 0, 0), spending 20 minutes. sqrt(3) ≈ 1.73205080756887729

	var emotions = make([]OCCEmotion, 0, len(e.Emotions))
	for _, emotion := range e.Emotions {
		emotion.Intensity *= math.Exp2(-seconds / halflife)
		if emotion.Intensity >= boundary {
			emotions = append(emotions, emotion)
		}
	}

	e.Emotions = emotions

	// 2. Move

	center := e.VirtualMoodCenter()
	if len(e.Emotions) == 0 {
		// returns to default
		center = e.Default.Vector3
		direction := center.Sub(e.Current.Position.Vector3)
		length := direction.Length()
		distance := math.Min(returnspeed*seconds, length)

		if length > 0 {
			step := direction.Scale(1 / length).Scale(distance)
			e.Current.Move(step)
		}
	} else {
		// returns to virtual mood center
		direction := center.Sub(e.Current.Position.Vector3)
		length := direction.Length()
		distance := math.Min(movespeed*seconds, length)

		if length > 0 {
			step := direction.Scale(1 / length).Scale(distance)
			e.Current.Move(step)
		}
	}
}
