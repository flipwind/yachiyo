package basic

import (
	"slices"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"
)

type BasicStorage struct {
	Histories   []event.Event
	nowSequence int64
}

func New() *BasicStorage {
	return &BasicStorage{
		Histories:   make([]event.Event, 0),
		nowSequence: 0,
	}
}

func (b *BasicStorage) Append(m event.Event) error {
	b.Histories = append(b.Histories, m)
	return nil
}

func (b *BasicStorage) ListAll() []event.Event {
	return slices.Clone(b.Histories)
}

func (b *BasicStorage) ListAfter(uuid uuid.UUID) []event.Event {
	var history_len = len(b.Histories)
	var index = history_len - 1
	var findflag = false
	for i := index; i >= 0; i-- {
		if b.Histories[i].GetMeta().ID == uuid {
			index = i
			findflag = true
			break
		}
	}

	if findflag == false {
		return slices.Clone(b.Histories)
	}

	return slices.Clone(b.Histories[index+1:])
}

func (b *BasicStorage) Close() error                                     { return nil }
func (b *BasicStorage) SaveCheckpoint(point history.Checkpoint) error    { return nil }
func (b *BasicStorage) SaveCompressNode(node history.CompressNode) error { return nil }
func (b *BasicStorage) NextSequence() int64 {
	b.nowSequence += 1
	return b.nowSequence
}
func (b *BasicStorage) NowSequence() int64 {
	return b.nowSequence
}
