package provider

import (
	"context"
	"yachiyo/yachiyo-core/history"
	"yachiyo/yachiyo-utils/logger"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var log = logger.New("Yachiyo.Provider")

type OpenAIProvider struct {
	client openai.Client
	model  string
}

func NewOpenAIProvider(baseUrl, apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseUrl),
		),
		model: model,
	}
}

func (p *OpenAIProvider) LLM() {}

func (p *OpenAIProvider) Gen(history []history.History) (string, error) {
	var OpenAIMessages []openai.ChatCompletionMessageParamUnion
	for _, m := range history {
		switch m.Role {
		case "system":
			OpenAIMessages = append(OpenAIMessages, openai.SystemMessage(m.Content))
		case "user":
			OpenAIMessages = append(OpenAIMessages, openai.UserMessage(m.Content))
		case "assistant":
			OpenAIMessages = append(OpenAIMessages, openai.AssistantMessage(m.Content))
		}
	}

	reply, err := p.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: OpenAIMessages,
		Model:    p.model,
	})

	if err != nil {
		return "", err
	}

	return reply.Choices[0].Message.Content, nil
}
