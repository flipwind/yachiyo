package onebot

type Message struct {
	MessageItems []MessageSegment
}

type MessageSegment struct {
	Type string		`json:"type"`
	Data any		`json:"data"`
}

type GroupMessageSend struct {
	GroupID int64
	Content string
}