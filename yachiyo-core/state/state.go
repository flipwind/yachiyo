package state

import "fmt"

type State struct {
	Socialization Urgency
	Interest Urgency
}

func NewState() State {
	return State{
		Socialization: Medium,
		Interest: Medium,
	}
}

func (s *State) String() string {
	return fmt.Sprintf(`<Socialization: %s, Interest: %s>`, s.Socialization.String(), s.Interest.String())
}

func (s *State) Prompt() string {
	result := ""

	var advice string
	switch s.Socialization{
	case Trivial, Low:
		advice = "Don't want to talk. Make conversation brief."
	case Medium:
		advice = "Common sense, depending on the topic."
	case High, Urgent:
		advice = "Be likely want to talk. Make conversation a bit longer."
	case Extreme:
		advice = "Want to talk more. Try to make conversation active."
	}
	result += fmt.Sprintf("Socialization: %s(%s), ", s.Socialization.String(), advice)

	switch s.Interest{
	case Trivial, Low:
		advice = "May don't like this topic or sender."
	case Medium, High, Urgent:
		advice = "Want to learn more about this topic."
	case Extreme:
		advice = "Active and talkative."
	}
	result += fmt.Sprintf("Interest: %s(%s)>", s.Interest.String(), advice)

	return result
}