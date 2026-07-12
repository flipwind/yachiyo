package basic

import (
	"time"
	"yachiyo/yachiyo-runtime/history"
)

type BasicStorage struct {
	Histories []history.History
}

func New() *BasicStorage{
	return &BasicStorage{
		Histories: make([]history.History, 0),
	}
}

func (b *BasicStorage) Remember(m history.History){
	b.Histories = append(b.Histories, m)
}

func (b *BasicStorage) ListAll() []history.History{
	return b.Histories
}

func (b *BasicStorage) GetLastActive() time.Time{
	if len(b.Histories)-1 < 0 {
		return time.Now()
	}
	return b.Histories[len(b.Histories)-1].Time
}