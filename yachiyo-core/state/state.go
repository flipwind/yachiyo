package state

import "fmt"

type State struct {
	Socialization Urgency
	Interest Urgency
	Attention Urgency
}

func New() *State {
	return &State{
		Socialization: Medium,
		Interest: Medium,
		Attention: Medium,
	}
}

func (s *State) String() string {
	return fmt.Sprintf(`<Socialization: %s, Interest: %s, Attention: %s>`, s.Socialization.String(), s.Interest.String(), s.Attention.String())
}