package state

import (
	"fmt"
	"reflect"
)

type State struct {
	SocialDesire Drive
	Interest Drive
}

func NewState() State {
	return State{
		SocialDesire: UrgencyDrive(Medium),
		Interest: UrgencyDrive(Medium),
	}
}

func (s *State) String() string {
	return fmt.Sprintf(`<SocialDesire: %s, Interest: %s>`, s.SocialDesire.String(), s.Interest.String())
}

func (s *State) Prompt() string {
	result := ""

	var advice string
	switch s.SocialDesire.Urgency(){
	case Trivial, Low:
		advice = "Don't want to talk. Make conversation brief. Avoid extend or introduce topic."
	case Medium:
		advice = "Common sense, depending on the topic."
	case High, Urgent:
		advice = "Be likely want to talk. Make conversation a bit longer."
	case Extreme:
		advice = "Want to talk more. Try to make conversation active."
	}
	result += fmt.Sprintf("SocialDesire: %s(%s), ", s.SocialDesire.String(), advice)

	switch s.Interest.Urgency(){
	case Trivial, Low:
		advice = "This topic is not attractive. Don't want to introduce new topic unless necessarily."
	case Medium, High, Urgent:
		advice = "Want to learn more about this topic."
	case Extreme:
		advice = "Active and talkative."
	}
	result += fmt.Sprintf("Interest: %s(%s)>", s.Interest.String(), advice)

	return result
}

func (s *State) Drives() []*Drive {
	v := reflect.ValueOf(s).Elem()
	
	drives := make([]*Drive, 0, v.NumField())
	for _, f := range v.Fields() {
		drives = append(drives, f.Addr().Interface().(*Drive))
	}

	return drives
}