package model

func (*RuntimeStateRequest) DataPack() {}

type RuntimeStateRequest struct{}

func (*RuntimeState) DataPack() {}

type RuntimeState struct {
	Timetick int64   `json:"timetick"`
	Energy   float64 `json:"energy"`
	Affect   struct {
		Octants   string       `json:"octants"`
		Pleasure  float64      `json:"pleasure"`
		Arousal   float64      `json:"arousal"`
		Dominance float64      `json:"dominance"`
		Actives   []OCCEmotion `json:"actives"`
	} `json:"affect"`
}

type OCCEmotion struct {
	Type      string  `json:"type"`
	Intensity float64 `json:"intensity"`
}
