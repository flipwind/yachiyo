package memory

type MemoryStorage interface {
	Remember(memory Memory)
	ListAll() []Memory
}

type Memory struct {
	Content string
}