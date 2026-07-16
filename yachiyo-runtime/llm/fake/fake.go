package fake

import (
	"fmt"
	"yachiyo/yachiyo-runtime/history"
)

type FakeLLM struct{}

func NewFakeLLM() *FakeLLM {
	return &FakeLLM{}
}

func (l *FakeLLM) LLM() {}

func (l *FakeLLM) Gen(history []history.History) (string, error) {
	return fmt.Sprintf("%v", history), nil
}