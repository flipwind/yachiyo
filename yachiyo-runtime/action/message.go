package action

type Message struct {
	Content  string
	Time     int64

	Address Address // The address of the message. e.g. onebot://group/12345
}

func (m *Message) action() {}