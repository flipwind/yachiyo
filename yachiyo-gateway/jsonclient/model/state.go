package model

func (*RuntimeStateRequest) DataPack(){}
type RuntimeStateRequest struct {}

func (*RuntimeState) DataPack(){}
type RuntimeState struct {
	State string `json:"state"`
}