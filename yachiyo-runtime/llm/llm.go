package llm

import (
	"context"
	"encoding/json"
)

type Client interface {
	Generate(ctx context.Context, req Request) (Response, error)
}

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type FinishReason string

const (
	FinishStop          FinishReason = "stop"
	FinishToolCall      FinishReason = "tool_calls"
	FinishLength        FinishReason = "length"
	FinishContentFilter FinishReason = "content_filter"
	FinishUnavailable   FinishReason = "insufficient_system_resource"
	FinishAborted       FinishReason = "aborted"
)

type ToolSpec struct {
	Name        string
	Description string
	Parameters  json.RawMessage //exhaustruct:optional
}

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type Message struct {
	Role    Role
	Content string

	// Set on an assistant message that requires toolcalls
	ToolCalls []ToolCall //exhaustruct:optional
	// Set on a callback tool message
	ToolCallID string //exhaustruct:optional
}

type Request struct {
	Messages []Message
	JSON     bool       // JSON Constraints
	Tools    []ToolSpec //exhaustruct:optional
}

type Response struct {
	FinishReason FinishReason
	Content      string
	ToolCalls    []ToolCall
}
