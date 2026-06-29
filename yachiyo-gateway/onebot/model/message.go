package onebotModel

type Message struct {
	MessageItems []MessageSegment
}

type MessageSegment struct {
	Type string		`json:"type"`
	Data any		`json:"data"`
}