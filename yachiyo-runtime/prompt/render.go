package prompt

import (
	"fmt"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/trigger"
)

func RenderUserMessage(m trigger.UserMessage) string {
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Timestamp.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}

func Drives(d state.Drives) string {
	result := ""

	var advice string
	switch d.SocialDesire.Urgency() {
	case state.Trivial, state.Low:
		advice = "Don't want to talk. Make conversation brief. Avoid extend or introduce topic."
	case state.Medium:
		advice = "Common sense, depending on the topic."
	case state.High, state.Urgent:
		advice = "Be likely want to talk. Make conversation a bit longer."
	case state.Extreme:
		advice = "Want to talk more. Try to make conversation active."
	}
	result += fmt.Sprintf("SocialDesire: %s(%s), ", d.SocialDesire.String(), advice)

	switch d.Interest.Urgency() {
	case state.Trivial, state.Low:
		advice = "This topic is not attractive. Don't want to introduce new topic unless necessarily."
	case state.Medium, state.High, state.Urgent:
		advice = "Want to learn more about this topic."
	case state.Extreme:
		advice = "Active and talkative."
	}
	result += fmt.Sprintf("Interest: %s(%s)>", d.Interest.String(), advice)

	return result
}

func Affect(cur state.PAD, emotion []state.OCCEmotion) string {
	moods := ""
	for _, mood := range emotion {
		moods += fmt.Sprintf("	- %s: %f\n", mood.Type.Name, mood.Intensity)
	}

	if moods == "" {
		moods = "	- Normal"
	}

	return fmt.Sprintf(`Mood Octants: %s
PAD Space: %s
Active Moods:
%s`, cur.Category(), cur.String(), moods)
}
