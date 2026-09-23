package llm

import (
	"yachiyo/yachiyo-runtime/prompt"
)

type LLM interface {
	LLM()
	Gen(history []prompt.Prompts) (string, error)
}