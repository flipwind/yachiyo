package basic

import (
	"uuid"
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

func (b *BasicStorage) ListAfter(uuid uuid.UUID) []history.History {
	var history_len = len(b.Histories)
	var index = history_len - 1
	var findflag = false
	for i := index; i >= 0; i-- {
		if b.Histories[i].GetID() == uuid {
			index = i
			findflag = true
			break
		}
	}

	if findflag == false {
		return b.Histories
	}

	return b.Histories[index+1:]
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
