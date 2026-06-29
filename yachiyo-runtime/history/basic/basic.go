package basic

import "yachiyo/yachiyo-runtime/history"

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