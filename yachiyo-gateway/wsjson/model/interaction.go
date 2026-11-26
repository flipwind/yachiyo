package model

func (_ *ClientMessage) DataPack() {}

type ClientMessage struct {
	Message string `json:"message"`
	Time    int64  `json:"time"`
	Author  string `json:"author,omitempty"` //exhaustruct:optional
	Path    string `json:"path,omitempty"`   //exhaustruct:optional
}

func (_ *RuntimeMessage) DataPack() {}

type RuntimeMessage struct {
	Reply        bool     `json:"reply"`
	Message      []string `json:"message"`
	IsInitiative bool     `json:"is_initiative"`
	Time         int64    `json:"time"`
	Path         string   `json:"path,omitempty"` //exhaustruct:optional
}

func (*GetRelativeMessageHistory) DataPack() {}

type GetRelativeMessageHistory struct{}

func (*RelativeMessageHistory) DataPack() {}

type RelativeMessageHistory struct {
	Messages []Envelope `json:"messages"`
}

func (*RuntimeError) DataPack() {}

type RuntimeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"` //exhaustruct:optional
}
