package llm

import (
	"yachiyo/yachiyo-runtime/prompt"
)

type LLM interface {
	LLM()
	Gen(history []prompt.Prompts, jsonEnabled bool) (answer string, err error, finish_reason string)
}
