package model

func (_ *Register) DataPack(){}
type Register struct {
	ClientType string `json:"client_type"`
	ClientName string `json:"client_name"`
	ClientID   string `json:"client_id"`
}

func (_ *RegisterSuccess) DataPack(){}
type RegisterSuccess struct {
	RuntimeName    string `json:"runtime_name"`
	RuntimeVersion string `json:"runtime_version"`
}

func (_ *RegisterError) DataPack(){}
type RegisterError struct {
	ErrorType string `json:"error_type"`
}

func (_ *HeartBeat) DataPack(){}
type HeartBeat struct{}

func (_ *HeartBeatRespond) DataPack(){}
type HeartBeatRespond struct{}

func (_ *Offline) DataPack(){}
type Offline struct{}
