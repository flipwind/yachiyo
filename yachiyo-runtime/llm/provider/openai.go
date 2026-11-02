package provider

import (
	"context"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/prompt"
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

func (p *OpenAIProvider) Gen(prompts []prompt.Prompts, jsonEnabled bool) (answer string, err error, finish_reason string) {
	var OpenAIMessages []openai.ChatCompletionMessageParamUnion
	for _, m := range prompts {
		switch p := m.(type) {
		case prompt.CompressPrompt:
			OpenAIMessages = append(OpenAIMessages, openai.UserMessage(p.Content))
		case prompt.SystemPrompt:
			OpenAIMessages = append(OpenAIMessages, openai.SystemMessage(p.Content))
		case prompt.UserPrompt:
			OpenAIMessages = append(OpenAIMessages, openai.UserMessage(p.Content))
		case prompt.AssistantPrompt:
			OpenAIMessages = append(OpenAIMessages, openai.AssistantMessage(p.Content))
		default:
			ylog.Warn("Unknown type of prompt: %T", p)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var format openai.ChatCompletionNewParamsResponseFormatUnion
	if jsonEnabled {
		format = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	} else {
		format = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &shared.ResponseFormatTextParam{},
		}
	}

	reply, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: OpenAIMessages,
		Model:    p.model,
		ResponseFormat: format,
	})

	if err != nil {
		return "", err, ""
	}

	if len(reply.Choices) == 0 {
		return "", fmt.Errorf("reply.choices is empty: %w", yerror.ErrLLMChoice), ""
	}

	return reply.Choices[0].Message.Content, nil, reply.Choices[0].FinishReason
}
