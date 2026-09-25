package history

import (
	"fmt"
	"slices"
	"time"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/initiative"
	"yachiyo/yachiyo-runtime/state"
)

type HistoryStorage interface {
	Append(memory History)
	ListAll() []History
}

type History interface {
	History()
}

type Snapshot struct {
	Time           time.Time // The time when snapshot was taken
	Emotion        state.Emotion
	State          state.State
	Factors        initiative.Factors
	LastActiveTime time.Time
}

type UserMessage struct {
	Snapshot Snapshot
	Content  string

	Author   string
	Platform string

	Time    time.Time
	Address address.Address
}

func (UserMessage) History() {}
func (m UserMessage) Render() string {
	// e.g. <Iroha>(2030.07.21 16:16:08/Client) Good evening Yachiyo~
	return fmt.Sprintf("<%s>(%s/%s) %s", m.Author, m.Time.Format("2006.01.02 15:04:05"), m.Platform, m.Content)
}

type AssistantMessage struct {
	Snapshot Snapshot
	Content  string

	Time      time.Time
	ToAddress address.Address
	Note      string

	Initiative bool
	Reply      bool
}

func (AssistantMessage) History() {}
func (m AssistantMessage) Render() string {
	// e.g. <Message initiative: False><Reply: True>(2030.07.21 16:16:09) Iroha! Suki~
	msg := fmt.Sprintf("<Message initiative: %t><Reply: %t>(%s) ", m.Initiative, m.Reply, m.Time.Format("2006.01.02 15:04:05"))
	if m.Reply {
		msg += m.Content
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
