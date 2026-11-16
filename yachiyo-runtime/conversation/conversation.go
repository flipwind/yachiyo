package conversation

import (
	"slices"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-shared/address"
)

type Summary struct {
	ID       uuid.UUID
	CreateAt time.Time
	Parent   uuid.UUID // uuid.Nil() if none
	RangeID  struct {
		From uuid.UUID
		To   uuid.UUID
	}
	Content string
}

type Window struct {
	Summary *Summary
	Events  []event.Event
}

func LastUserAddress(events []event.Event) (address.Address, bool) {
	for _, e := range slices.Backward(events) {
		if item, ok := e.(event.UserMessage); ok {
			return item.Address, true
		}
	}

	return address.Address{}, false
}

type Conversation struct {
	events   []event.Event
	summary  *Summary
	startAt  int
	note     string
	sequence int64
}

func Restore(events []event.Event, summary *Summary, note string, sequence int64) *Conversation {
	conv := &Conversation{
		events:   slices.Clone(events),
		summary:  nil,
		startAt:  0,
		note:     note,
		sequence: sequence,
	}

	for _, e := range conv.events {
		conv.sequence = max(conv.sequence, e.GetMeta().Sequence)
	}

	if summary != nil {
		conv.SetSummary(*summary)
	}

	return conv
}

func (c *Conversation) Append(e event.Event) {
	c.events = append(c.events, e)
}

func (c *Conversation) Sequence() int64 {
	return c.sequence
}

func (c *Conversation) NextSequence() int64 {
	c.sequence += 1
	return c.sequence
}

func (c *Conversation) Window() Window {
	var s *Summary
	if c.summary != nil {
		cp := *c.summary
		s = &cp
	}

	return Window{
		Summary: s,
		Events:  slices.Clone(c.events[c.startAt:]),
	}
}

func (c *Conversation) Events() []event.Event {
	return slices.Clone(c.events)
}

func (c *Conversation) Note() string {
	return c.note
}

func (c *Conversation) SetSummary(s Summary) {
	c.summary = &s

	c.startAt = 0
	for i := len(c.events) - 1; i >= 0; i-- {
		if c.events[i].GetMeta().ID == s.RangeID.To {
			c.startAt = i + 1
			break
		}
	}
}

func (c *Conversation) SetNote(note string) {
	c.note = note
}
