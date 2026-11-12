package tool

import "yachiyo/yachiyo-runtime/llm"

type Tool struct {
	Spec    llm.ToolSpec
	Execute func(arguments string) (string, error)
}
