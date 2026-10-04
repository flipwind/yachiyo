package action

import "yachiyo/yachiyo-runtime/address"

type ErrorCode string

const (
	ErrLLMChoiceEmpty ErrorCode = "llm_choice_empty"
	ErrLLMGenerate    ErrorCode = "llm_generate"
	ErrMessageProcess ErrorCode = "message_process"
)

type Error struct {
	Code    ErrorCode
	Message string
	Address address.Address
}

func (e *Error) action() {}
func (e *Error) GetAddress() address.Address {
	return e.Address
}
