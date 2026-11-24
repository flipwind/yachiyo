package model

import (
	"encoding/json"
	"regexp"
)

type ToolItem struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9_]{1,32}$`)

func (t ToolItem) ValidName() bool {
	return validName.MatchString(t.Name)
}

type UpdateTools struct {
	Capability []string   `json:"capability"`
	Tools      []ToolItem `json:"tools"`
}

func (*UpdateTools) DataPack() {}

type UpdateToolsFeedback struct {
	Status           string            `json:"status"`
	RejectCapability []string          `json:"reject_capability"`
	RejectTools      map[string]string `json:"reject_tools"`
}

func (*UpdateToolsFeedback) DataPack() {}

type ToolCall struct {
	ToolName  string          `json:"tool_name"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

func (*ToolCall) DataPack() {}

type ToolCallResult struct {
	RequestID string `json:"request_id"`
	OK        bool   `json:"ok"`
	Result    string `json:"result"`
}

func (*ToolCallResult) DataPack() {}
