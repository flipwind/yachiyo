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

func (b *BasicStorage) GetLastUserHistory() history.History{
	if len(b.Histories)-1 < 0 {
		return history.History{}
	}

	var h = history.History{}
	for i := len(b.Histories) - 1; i >= 0; i -- {
		if b.Histories[i].Role == "user" {
			h = b.Histories[i]
			break
		}
	}
	return h
}