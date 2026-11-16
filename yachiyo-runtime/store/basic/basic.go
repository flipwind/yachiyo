package basic

import (
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/store"
)

var _ store.Store = (*Storage)(nil)

type Storage struct {
	Histories []event.Event
}

func (s *Storage) Load() (store.Restored, error) {
	return store.Restored{}, nil
}

func (s *Storage) Append(e event.Event) error {
	s.Histories = append(s.Histories, e)
	return nil
}

func (s *Storage) SaveCheckpoint(c store.Checkpoint) error        { return nil }
func (s *Storage) SaveSummary(summary conversation.Summary) error { return nil }
func (s *Storage) Close() error                                   { return nil }
