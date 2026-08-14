package model

type ClientMessage struct {
	Message string `json:"message"`
}

type RuntimeMessage struct {
	Message      string `json:"message"`
	IsInitiative bool   `json:"is_initiative"`
}
