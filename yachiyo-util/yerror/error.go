package yerror

import (
	"errors"
)

var (
	ErrFieldRequired = errors.New("field required")
	ErrFieldInvalid = errors.New("field invalid")
	ErrFieldIncomplete = errors.New("field incomplete")

	ErrLLMChoice = errors.New("LLM didn't give a choice")
	ErrLLMGenerate = errors.New("LLM generating failed")
)