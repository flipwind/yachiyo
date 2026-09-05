package model

func (_ *ClientMessage) DataPack(){}
type ClientMessage struct {
	Message string `json:"message"`
}

func (_ *RuntimeMessage) DataPack(){}
type RuntimeMessage struct {
	Reply        bool 	`json:"reply"`
	Message      string `json:"message"`
	IsInitiative bool   `json:"is_initiative"`
}
