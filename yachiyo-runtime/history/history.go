package history

type HistoryStorage interface {
	Remember(memory History)
	ListAll() []History
	GetLastActive() string
}

type History struct {
	Role string
	Content string
	Time string
}