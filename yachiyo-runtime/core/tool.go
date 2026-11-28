package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/hub"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/tool"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
)

func getBuildinTools(snap Snapshot, cause prompt.Cause, sent *[]sentMessage, hub hub.Hub) []tool.Tool {
	builtinTools := make([]tool.Tool, 0)

	test := tool.Tool{
		Spec: llm.ToolSpec{
			Description: "Return a fixed string",
			Name:        "get_a_string",
		},
		Execute: func(payload string) (string, error) {
			return "Hello yachiyo!", nil
		},
	}

	get_affect := tool.Tool{
		Spec: llm.ToolSpec{
			Description: "Return the Runtime Affect string",
			Name:        "get_runtime_affect",
		},
		Execute: func(payload string) (string, error) {
			return prompt.Affect(snap.State.Mood, snap.State.Emotions), nil
		},
	}

	availableAddr := tool.Tool{
		Spec: llm.ToolSpec{
			Description: "Query currently online addresses that support a given capability, like message.text",
			Name:        "available_address",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"capability": {"type":"string", "description":"Capability name, now only message.text"}
				},
				"required": ["capability"]
			}`),
		},
		Execute: func(payload string) (string, error) {
			var args struct {
				Capability string `json:"capability"`
			}
			if err := json.Unmarshal([]byte(payload), &args); err != nil {
				return "", err
			}
			addrs, err := hub.AvailableAddresses(args.Capability)
			if err != nil {
				return "", err
			}

			li := make([]string, 0)
			if len(addrs) == 0 {
				return "no available address", nil
			}
			for _, a := range addrs {
				li = append(li, a.Content)
			}

			byt, _ := json.Marshal(li)
			return string(byt), nil
		},
	}

	commitMessage := tool.Tool{
		Spec: llm.ToolSpec{
			Name:        "commit_message",
			Description: "Send text messages to a target address.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"address":  {"type": "string", "description": "Target address which own ability message.text"},
					"messages": {"type": "array", "items": {"type": "string"}, "description": "List of message bubbles to send"}
				},
				"required": ["address", "messages"]
			}`),
		},
		Execute: func(payload string) (string, error) {
			var args struct {
				Address  string   `json:"address"`
				Messages []string `json:"messages"`
			}
			if err := json.Unmarshal([]byte(payload), &args); err != nil {
				return "", err
			}

			if len(args.Messages) == 0 {
				return "", errors.New("messages cannot be empty")
			}

			now := time.Now()
			addr := address.New(args.Address)
			err := hub.SendMessage(action.AssistantMessage{
				Address:    addr,
				Content:    args.Messages,
				Timestamp:  now,
				Initiative: cause == prompt.Initiative,
			})

			if err != nil {
				return "", err
			}

			*sent = append(*sent, sentMessage{
				Address:   addr,
				Content:   args.Messages,
				Timestamp: now,
			})

			return "Message sent successfully.", nil
		},
	}

	builtinTools = append(builtinTools, test, get_affect, availableAddr, commitMessage)

	return builtinTools
}

func getTools(ctx context.Context, snap Snapshot, cause prompt.Cause, sent *[]sentMessage, hub hub.Hub) []tool.Tool {
	tools := make([]tool.Tool, 0)

	builtinTools := getBuildinTools(snap, cause, sent, hub)
	tools = append(tools, builtinTools...)
	for _, t := range hub.ClientTools() {
		tools = append(tools, tool.Tool{
			Spec: llm.ToolSpec{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
			Execute: func(payload string) (string, error) {
				reqID := uuid.NewV7()
				return hub.UseTool(ctx, t.Address, reqID.String(), t.Name, payload)
			},
		})
	}

	return tools
}
