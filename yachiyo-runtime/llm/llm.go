package llm

import (
	"yachiyo/yachiyo-runtime/prompt"
)

type LLM interface {
	LLM()
	Gen(history []prompt.Prompts, jsonEnabled bool, tools ...prompt.Tool) (result Result, err error)
}

type Result struct {
	FinishReason string
	Content      string
	ToolCall     []prompt.ToolCall
}
