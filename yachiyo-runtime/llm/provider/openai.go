package provider

import (
	"context"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/yerror"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
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

func toMessages(msgs []llm.Message) (openAIMessages []openai.ChatCompletionMessageParamUnion) {
	for _, msg := range msgs {
		switch msg.Role {
		case llm.RoleSystem:
			openAIMessages = append(openAIMessages, openai.SystemMessage(msg.Content))
		case llm.RoleUser:
			openAIMessages = append(openAIMessages, openai.UserMessage(msg.Content))
		case llm.RoleAssistant:
			m := openai.AssistantMessage(msg.Content)

			for _, tool := range msg.ToolCalls {
				m.OfAssistant.ToolCalls = append(m.OfAssistant.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: tool.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      tool.Name,
							Arguments: tool.Arguments,
						},
					},
				})
			}

			openAIMessages = append(openAIMessages, m)
		case llm.RoleTool:
			openAIMessages = append(openAIMessages, openai.ToolMessage(msg.Content, msg.ToolCallID))
		default:
			ylog.Warn("Unknown type of prompt: %T", msg)
		}
	}

	return openAIMessages
}

func toTools(tools []llm.ToolSpec) (toolParams []openai.ChatCompletionToolUnionParam) {
	for _, t := range tools {
		properties := make(map[string]any, len(t.Parameters))
		required := make([]string, 0)

		for name, p := range t.Parameters {
			properties[name] = map[string]any{
				"type":        p.Type,
				"description": p.Description,
			}
			if p.Required {
				required = append(required, name)
			}
		}

		toolParams = append(toolParams, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        t.Name,
					Description: param.NewOpt(t.Description),
					Parameters: shared.FunctionParameters{
						"type":       "object",
						"properties": properties,
						"required":   required,
					},
				},
			},
		})
	}

	return toolParams
}

func (p *OpenAIProvider) Generate(ctx context.Context, req llm.Request) (llm.Response, error) {
	var result llm.Response
	OpenAIMessages := toMessages(req.Messages)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var format openai.ChatCompletionNewParamsResponseFormatUnion
	if req.JSON {
		format = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	} else {
		format = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &shared.ResponseFormatTextParam{},
		}
	}

	toolparam := toTools(req.Tools)

	reply, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:       OpenAIMessages,
		Model:          p.model,
		ResponseFormat: format,
		Tools:          toolparam,
	})

	if err != nil {
		return result, err
	}

	if len(reply.Choices) == 0 {
		return result, fmt.Errorf("reply.choices is empty: %w", yerror.ErrLLMChoice)
	}

	var rep = reply.Choices[0]

	toolcalls := make([]llm.ToolCall, 0)
	for _, t := range rep.Message.ToolCalls {
		toolcalls = append(toolcalls, llm.ToolCall{
			ID:        t.ID,
			Name:      t.Function.Name,
			Arguments: t.Function.Arguments,
		})
	}

	result = llm.Response{
		FinishReason: llm.FinishReason(rep.FinishReason),
		Content:      rep.Message.Content,
		ToolCalls:    toolcalls,
	}

	return result, nil
}
