package provider

import (
	"context"
	"time"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/yerror"

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
		case "user", "user/runtime":
			OpenAIMessages = append(OpenAIMessages, openai.UserMessage(m.Content))
		case "assistant":
			OpenAIMessages = append(OpenAIMessages, openai.AssistantMessage(m.Content))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	reply, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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

	if len(reply.Choices) == 0 {
		ylog.Error("Reply's choices is empty.")
		return "", yerror.TypeMissing("reply.choices")
	}

	return reply.Choices[0].Message.Content, nil
}
