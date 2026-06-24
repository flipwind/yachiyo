package llm

import "yachiyo/yachiyo-core/history"

type LLM interface {
	LLM()
	Gen(history []history.History) string
}