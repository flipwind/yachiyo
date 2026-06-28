package onebot

import "yachiyo/yachiyo-core/event"

type AdapterChannel struct {
	ToServer chan event.UserMessageEvent
	ToClient chan GroupMessageSend
}

func NewAdapterChannel() *AdapterChannel {
	return &AdapterChannel{
		ToServer: make(chan event.UserMessageEvent),
		ToClient: make(chan GroupMessageSend),
	}
}