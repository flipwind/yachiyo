package onebotModel

type Event struct {
	Time int64		`json:"time"`
	SelfID int64	`json:"self_id"`
	PostType string	`json:"post_type"`
}

type GroupMessageEvent struct {
	Event
	MessageType string 	`json:"message_type"`
	SubType string		`json:"sub_type"`
	MessageID int64		`json:"message_id"`
	GroupID int64		`json:"group_id"`
	UserID int64		`json:"user_id"`
	Message any			`json:"message"`		// TODO: any to universal
	RawMessage string	`json:"raw_message"`
	Sender User			`json:"sender"`
}