package history

import (
	"fmt"
	"slices"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
)

type HistoryStorage interface {
	Append(memory History) error
	ListAll() []History
	Close() error
	SaveCheckpoint(point Checkpoint) error
	NextSequence() int64
	NowSequence() int64
}

type History interface {
	History()
	GetSequence() int64
}

type Event struct {
	ID        uuid.UUID
	Sequence  int64
	Timestamp time.Time
	Snapshot  Snapshot
	Address   address.Address
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

type Snapshot struct {
	Time           time.Time // The time when snapshot was taken
	Affect         state.Engine
	State          state.State
	Energy         float64
	LastActiveTime time.Time
}

type UserMessage struct {
	Event
	Content  string
	Author   string
	Platform string
}

func (UserMessage) History() {}

func (m UserMessage) GetSequence() int64 { return m.Sequence }
func (m UserMessage) Render() string {
	// e.g. <Iroha>(2030.07.21 16:16:08/Client) Good evening Yachiyo~
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Timestamp.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}

// To AssistantMessage, field `Address` means where the address the message would send to.
type AssistantMessage struct {
	Event
	Content []string

	Initiative bool
	Reply      bool
}

func (AssistantMessage) History() {}

func (m AssistantMessage) GetSequence() int64 { return m.Sequence }
func (m AssistantMessage) Render() string {
	// e.g. <Message initiative: False><Reply: True>(2030.07.21 16:16:09) Iroha! Suki~
	msg := fmt.Sprintf("<Message initiative: %t><Reply: %t>(%s) ", m.Initiative, m.Reply, m.Timestamp.Format("2006.01.02 15:04:05"))
	if m.Reply {
		for _, message := range m.Content {
			msg += fmt.Sprintf("`%s` ", message)
		}
	} else {
		msg += fmt.Sprintf("<Reason: %s>", m.Content)
	}

	return msg
}

func GetLastUserAddress(hist []History) (address.Address, bool) {
	for _, h := range slices.Backward(hist) {
		if item, ok := h.(UserMessage); ok {
			return item.Address, true
		}
	}

	return address.Address{}, false
}
