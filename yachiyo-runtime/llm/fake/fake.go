package fake

import (
	"fmt"
	"yachiyo/yachiyo-runtime/prompt"
)

type FakeLLM struct{}

func NewFakeLLM() *FakeLLM {
	return &FakeLLM{}
}

func (l *FakeLLM) LLM() {}

func (l *FakeLLM) Gen(history []prompt.Prompts) (string, error) {
	return fmt.Sprintf("%v", history), nil
}