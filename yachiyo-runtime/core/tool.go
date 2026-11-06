package core

import "yachiyo/yachiyo-runtime/prompt"

func (c *Core) GetBuiltinTools() []prompt.Tool {
	builtinTools := make([]prompt.Tool, 0)

	test := prompt.Tool{
		Description: "Return a fixed string",
		Name:        "get_a_string",
		Execute: func(payload string) (string, error) {
			return "Hello yachiyo!", nil
		},
	}

	get_affect := prompt.Tool{
		Description: "Return the Runtime Affect string",
		Name:        "get_runtime_affect",
		Execute: func(payload string) (string, error) {
			return c.Affect.String(), nil
		},
	}

	builtinTools = append(builtinTools, test, get_affect)

	return builtinTools
}

func (c *Core) GetTools() []prompt.Tool {
	tools := make([]prompt.Tool, 0)

	tools = append(tools, c.GetBuiltinTools()...)

	// implement third-party tools, from clients or other mcps.

	return tools
}
