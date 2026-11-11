package history

import (
	"slices"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-shared/address"
)

type HistoryStorage interface {
	Append(memory event.History) error
	ListAll() []event.History
	ListAfter(id uuid.UUID) []event.History // returns history[indexOfID+1:]
	Close() error
	SaveCheckpoint(point Checkpoint) error
	SaveCompressNode(node CompressNode) error
	NextSequence() int64
	NowSequence() int64
}

type Checkpoint struct {
	Exist          bool
	StateAt        time.Time
	State          state.State
	Determination  state.Determination
	Energy         initiative.Energy
	Affect         state.Engine
	Note           string
	LastActiveTime time.Time
	NowSequence    int64
}

type RestoreResult struct {
	Checkpoint Checkpoint
	LatestNode CompressNode
}

func GetLastUserAddress(hist []event.History) (address.Address, bool) {
	for _, h := range slices.Backward(hist) {
		if item, ok := h.(event.UserMessage); ok {
			return item.Address, true
		}
	}

	return address.Address{}, false
}
