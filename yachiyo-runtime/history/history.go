package history

import (
	"time"
	"yachiyo/yachiyo-runtime/address"
)

type HistoryStorage interface {
	Remember(memory History)
	ListAll() []History
	GetLastActive() time.Time
	GetLastUserHistory() History
}

type History struct {
	Role    string
	Content string
	Time    time.Time
	Address address.Address
}