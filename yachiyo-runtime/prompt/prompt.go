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
	Content string
}

func (AssistantPrompt) Prompts() {}
