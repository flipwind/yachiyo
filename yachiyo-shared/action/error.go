package action

type ErrorCode string

const (
	ErrLLMChoiceEmpty ErrorCode = "llm_choice_empty"
	ErrLLMGenerate    ErrorCode = "llm_generate"
	ErrMessageProcess ErrorCode = "message_process"
	ErrMessageStore   ErrorCode = "message_store"
)

type Error struct {
	BaseAction
	Code    ErrorCode
	Message string
}

func (*Error) action() {}
