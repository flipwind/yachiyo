package llm

type LLM interface{
	LLM()
	Gen(prompt string) string
}