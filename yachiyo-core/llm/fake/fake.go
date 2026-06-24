package fake

import (
	"fmt"
	"yachiyo/yachiyo-core/history"
)

type FakeLLM struct{}

func NewFakeLLM() *FakeLLM {
	return &FakeLLM{}
}

func (l *FakeLLM) LLM() {}

func (l *FakeLLM) Gen(history []history.History) string {
	return fmt.Sprintf("%v", history)
}