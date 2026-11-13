package core

import (
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/tool"
)

func (c *Core) GetBuiltinTools() []tool.Tool {
	builtinTools := make([]tool.Tool, 0)

	snap := c.snapshot()

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

	builtinTools = append(builtinTools, test, get_affect)

	return builtinTools
}

func (c *Core) GetTools() []tool.Tool {
	tools := make([]tool.Tool, 0)

	tools = append(tools, c.GetBuiltinTools()...)

	// implement third-party tools, from clients or other mcps.

	return tools
}
