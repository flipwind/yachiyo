package basic

import (
	"yachiyo/yachiyo-runtime/history"
)

type BasicStorage struct {
	Histories   []history.History
	nowSequence int64
}

func New() *BasicStorage {
	return &BasicStorage{
		Histories:   make([]history.History, 0),
		nowSequence: 0,
	}
}

func (b *BasicStorage) Append(m history.History) error {
	b.Histories = append(b.Histories, m)
	return nil
}

func (b *BasicStorage) ListAll() []history.History {
	return b.Histories
}

func (b *BasicStorage) Close() error                                  { return nil }
func (b *BasicStorage) SaveCheckpoint(point history.Checkpoint) error { return nil }
func (b *BasicStorage) NextSequence() int64 {
	b.nowSequence += 1
	return b.nowSequence
}
func (b *BasicStorage) NowSequence() int64 {
	return b.nowSequence
}
