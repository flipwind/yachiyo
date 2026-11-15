package tool

import (
	"fmt"
	"yachiyo/yachiyo-runtime/llm"
)

type Tool struct {
	Spec    llm.ToolSpec
	Execute func(arguments string) (string, error)
}

type Set []Tool

func (s Set) Specs() []llm.ToolSpec {
	var list []llm.ToolSpec = make([]llm.ToolSpec, 0, len(s))
	for _, t := range s {
		list = append(list, t.Spec)
	}
	return list
}

func (s Set) Call(name, arguments string) string {
	for _, t := range s {
		if t.Spec.Name == name && t.Execute != nil {
			result, err := t.Execute(arguments)
			if err != nil {
				return fmt.Sprintf("Error when calling tool <%s>: %v", name, err)
			}
			return result
		}
	}

	return fmt.Sprintf("There's not a tool called <%s>", name)
}
