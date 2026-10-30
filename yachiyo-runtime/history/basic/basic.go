package basic

import (
	"yachiyo/yachiyo-runtime/history"
)

type BasicStorage struct {
	Histories []history.History
}

func New() *BasicStorage {
	return &BasicStorage{
		Histories: make([]history.History, 0),
	}
}

func (b *BasicStorage) Append(m history.History) error {
	b.Histories = append(b.Histories, m)
	return nil
}

func (b *BasicStorage) ListAll() []history.History {
	return b.Histories
}

func (b *BasicStorage) Close() error                              { return nil }
func (b *BasicStorage) SaveCheckpoint(point history.Checkpoint) error { return nil }
