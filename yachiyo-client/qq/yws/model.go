package yws

import "encoding/json"

type Envelope struct {
	Category string          `json:"category"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
}

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
	ErrorType string `json:"error_type"`
}

type HeartBeat struct{}

type HeartBeatRespond struct{}

type Offline struct{}

type ClientMessage struct {
	Message string `json:"message"`
	Time    int64  `json:"time"`
	Author  string `json:"author,omitempty"`
	Path    string `json:"path,omitempty"`
}

type RuntimeMessage struct {
	Reply        bool     `json:"reply"`
	Message      []string `json:"message"`
	IsInitiative bool     `json:"is_initiative"`
	Time         int64    `json:"time"`
	Path         string   `json:"path,omitempty"`
}

type RuntimeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type ToolItem struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type UpdateTools struct {
	Capability []string   `json:"capability"`
	Tools      []ToolItem `json:"tools"`
}

type UpdateToolsFeedback struct {
	Status           string            `json:"status"`
	RejectCapability []string          `json:"reject_capability"`
	RejectTools      map[string]string `json:"reject_tools"`
}

type ToolCall struct {
	ToolName  string          `json:"tool_name"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type ToolCallResult struct {
	RequestID string `json:"request_id"`
	OK        bool   `json:"ok"`
	Result    string `json:"result"`
}
