package basic

import "yachiyo/yachiyo-core/memory"

type BasicStorage struct {
	Memories []memory.Memory
}

func New() *BasicStorage{
	return &BasicStorage{
		Memories: make([]memory.Memory, 0),
	}
}

func (b *BasicStorage) Remember(m memory.Memory){
	b.Memories = append(b.Memories, m)
}

func (b *BasicStorage) ListAll() []memory.Memory{
	return b.Memories
}