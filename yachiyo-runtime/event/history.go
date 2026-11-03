package event

type History interface {
	Event
	history()
}