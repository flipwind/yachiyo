package provider

import (
	"context"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-util/logger"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

var ylog = logger.New("Yachiyo.Provider")

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
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		},
	})

	// JSON_object mode is unreliable(!!!!!!!!!) on deepseek-v4.

	if err != nil {
		return "", err
	}

	return reply.Choices[0].Message.Content, nil
}
