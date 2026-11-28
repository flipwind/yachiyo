package prompt

import (
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/trigger"
)

const timeLayout = "2006.01.02 15:04:05"

func userMessage(author, platform, content string, address address.Address, t time.Time) string {
	// e.g. <Iroha>(2030.07.21 16:16:08/Client) Good evening Yachiyo~
	return fmt.Sprintf("<%s>(%s/%s, %s) %s", author, t.Format(timeLayout), platform, address.Content, content)
}

func RenderUserMessage(m trigger.UserMessage) string {
	return userMessage(m.Author, m.Platform, m.Content, m.Address, m.Timestamp)
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

func UserMessage(m event.UserMessage) string {
	return userMessage(m.Author, m.Platform, m.Content, m.Address, m.Timestamp)
}

func AssistantMessage(m event.AssistantMessage, nickname string) string {
	// e.g. <Yachiyo>(2030.07.21 16:16:09 -> wsjson://testclient/private/1, passive) `Iroha!` `Suki~`
	mode := "passive"
	if m.Initiative {
		mode = "initiative"
	}
	target := ""
	if m.Address.Content != "" {
		target = " -> " + m.Address.Content
	}

	msg := fmt.Sprintf("<%s>(%s%s, %s)", nickname, m.Timestamp.Format(timeLayout), target, mode)
	for _, message := range m.Content {
		msg += fmt.Sprintf(" `%s`", message)
	}
	return msg
}

func Event(e event.Event, nickname string) string {
	switch m := e.(type) {
	case event.UserMessage:
		return UserMessage(m)
	case event.AssistantMessage:
		return AssistantMessage(m, nickname)
	default:
		ylog.Error("Unknown event: %T", m)
		return ""
	}
}

func Summary(s conversation.Summary) string {
	return fmt.Sprintf(`<CompressNode>[Compressed at %s]
%s`, s.CreateAt.Format(timeLayout), s.Content)
}

func JSONReminder() llm.Message {
	return llm.Message{
		Role:    llm.RoleSystem,
		Content: "<PROCESS HINT> JSON MODE IS ENABLED. YOU MUST FOLLOW THE OUTPUT ROLE.",
	}
}
