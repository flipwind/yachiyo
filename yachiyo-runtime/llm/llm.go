package llm

import "yachiyo/yachiyo-runtime/history"

type LLM interface {
	LLM()
	Gen(history []history.History) (string, error)
}