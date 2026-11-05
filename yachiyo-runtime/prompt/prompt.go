package prompt

type Prompts interface {
	Prompts()
}

type SystemPrompt struct {
	Content string
}

func (SystemPrompt) Prompts() {}

type UserPrompt struct {
	Content string
}

func (UserPrompt) Prompts() {}

type AssistantPrompt struct {
	Content     string
	UseToolcall []ToolCall
}

func (AssistantPrompt) Prompts() {}

type CompressPrompt struct {
	Content string
}

func (CompressPrompt) Prompts() {}

type ToolPrompt struct {
	Content string
	CallID  string
}

func (ToolPrompt) Prompts() {}
