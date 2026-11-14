package history

import (
	"slices"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/address"
)

type HistoryStorage interface {
	Append(memory event.Event) error

	ListAll() []event.Event	// returns the *copy* of history
	ListAfter(id uuid.UUID) []event.Event // returns the *copy* of history[indexOfID+1:]

	Close() error
	SaveCheckpoint(point Checkpoint) error
	SaveCompressNode(node CompressNode) error
	NextSequence() int64
	NowSequence() int64
}

type Checkpoint struct {
	Exist       bool
	StateAt     time.Time
	State       state.Snapshot
	Note        string
	NowSequence int64
}

type RestoreResult struct {
	Checkpoint Checkpoint
	LatestNode CompressNode
}

func GetLastUserAddress(hist []event.Event) (address.Address, bool) {
	for _, h := range slices.Backward(hist) {
		if item, ok := h.(event.UserMessage); ok {
			return item.Address, true
		}
	}

	return address.Address{}, false
}
