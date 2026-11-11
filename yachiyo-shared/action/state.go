package action

type OCCEmotion struct {
	Type      string
	Intensity float64
}

type Affect struct {
	Octants   string
	Pleasure  float64
	Arousal   float64
	Dominance float64
	Actives   []OCCEmotion
}

type RuntimeState struct {
	BaseAction
	Timetick int64
	Energy   float64
	Affect   Affect
}

func (*RuntimeState) action() {}
