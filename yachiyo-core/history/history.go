package history

type HistoryStorage interface {
	Remember(memory History)
	ListAll() []History
}

type History struct {
	Role string
	Content string
}