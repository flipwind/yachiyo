package history

import "time"

type HistoryStorage interface {
	Remember(memory History)
	ListAll() []History
	GetLastActive() time.Time
}

type History struct {
	Role    string
	Content string
	Time    time.Time
}