package llmprovider

import (
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3"

	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-utils/logger"
	"context"
)

var sourcename string = "Yachiyo.Server"

type OpenAIProvider struct {
	client openai.Client
	model string
}

func CreateOpenAIProvider(baseUrl, apiKey, model string) *OpenAIProvider{
	return &OpenAIProvider{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseUrl),
		),
		model: model,
	}
}

func (provider *OpenAIProvider) ChatStream(ctx context.Context, messages []chat.Message) (<-chan string, error){
	output := make(chan string)

	var OpenAIMessages []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		switch m.Role {
		case "system": OpenAIMessages = append(OpenAIMessages, openai.SystemMessage(m.Content))
		case "user": OpenAIMessages = append(OpenAIMessages, openai.UserMessage(m.Content))
		case "assistant": OpenAIMessages = append(OpenAIMessages, openai.AssistantMessage(m.Content))
		}
	}

	stream := provider.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: OpenAIMessages,
		Model: provider.model,
	})

	go func(){
		defer close(output)

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				output <- chunk.Choices[0].Delta.Content
			}
		}

		if err := stream.Err(); err != nil {
			logger.Error(sourcename, "Streaming break: %v", err)
		}
	}()

	return output, nil
}