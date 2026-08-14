package model

type Register struct {
	ClientType string `json:"client_type"`
	ClientName string `json:"client_name"`
	ClientID   string `json:"client_id"`
}

type RegisterSuccess struct {
	RuntimeName    string `json:"runtime_name"`
	RuntimeVersion string `json:"runtime_version"`
}

type RegisterError struct {
	ErrorType string `json:"register_error"`
}

type HeartBeat struct{}

type HeartBeatRespond struct{}

type Offline struct{}
