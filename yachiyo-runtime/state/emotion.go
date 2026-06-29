package state

import "fmt"

type Emotion struct {
	Type    string
	Urgency Urgency
}

func NewEmotion() Emotion {
	return Emotion{
		Type:    "Happy",
		Urgency: High,
	}
}

func (e *Emotion) String() string {
	return fmt.Sprintf("<@%s: %s>", e.Type, e.Urgency.String())
}