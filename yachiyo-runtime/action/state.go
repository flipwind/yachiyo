package action

import "yachiyo/yachiyo-runtime/address"

type OCCEmotion struct {
	Type      string
	Intensity float64
}

type RuntimeState struct {
	Timetick int64
	Energy   float64
	Affect   struct {
		Octants   string
		Pleasure  float64
		Arousal   float64
		Dominance float64
		Actives    []OCCEmotion
	}

	Address address.Address // The address of the message. e.g. onebot://group/12345
}

func (m *RuntimeState) action() {}
func (m *RuntimeState) GetAddress() address.Address {
	return m.Address
}
