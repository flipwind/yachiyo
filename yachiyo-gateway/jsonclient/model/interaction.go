package model

func (_ *ClientMessage) DataPack() {}

type ClientMessage struct {
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

func (_ *RuntimeMessage) DataPack() {}

type RuntimeMessage struct {
	Reply        bool   `json:"reply"`
	Message      string `json:"message"`
	IsInitiative bool   `json:"is_initiative"`
	Time         int64  `json:"time"`
}

func (*GetRelativeMessageHistory) DataPack() {}

type GetRelativeMessageHistory struct{}

func (*RelativeMessageHistory) DataPack() {}

type RelativeMessageHistory struct {
	Messages []Envelope `json:"messages"`
}
