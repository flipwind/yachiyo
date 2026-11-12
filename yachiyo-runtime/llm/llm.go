package llm

import (
	"context"
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
	Parameters  map[string]ToolParameter //exhaustruct:optional
}

type ToolParameter struct {
	Type        string
	Description string
	Required    bool
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
