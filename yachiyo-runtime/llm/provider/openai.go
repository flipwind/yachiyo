package provider

import (
	"context"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
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

func (p *OpenAIProvider) LLM() {}

func (p *OpenAIProvider) Gen(prompts []prompt.Prompts, jsonEnabled bool, tools ...prompt.Tool) (result llm.Result, err error) {
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
			msg := openai.AssistantMessage(p.Content)

			for _, tool := range p.UseToolcall {
				msg.OfAssistant.ToolCalls = append(msg.OfAssistant.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: tool.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      tool.Name,
							Arguments: tool.Payload,
						},
					},
				})
			}

			OpenAIMessages = append(OpenAIMessages, msg)
		case prompt.ToolPrompt:
			OpenAIMessages = append(OpenAIMessages, openai.ToolMessage(p.Content, p.CallID))
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

	toolparam := make([]openai.ChatCompletionToolUnionParam, 0)
	for _, t := range tools {
		properties := make(map[string]any, len(tools))
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

		toolparam = append(toolparam, openai.ChatCompletionToolUnionParam{
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

	toolcalls := make([]prompt.ToolCall, 0)
	for _, t := range rep.Message.ToolCalls {
		toolcalls = append(toolcalls, prompt.ToolCall{
			ID:      t.ID,
			Name:    t.Function.Name,
			Payload: t.Function.Arguments,
		})
	}

	result = llm.Result{
		FinishReason: rep.FinishReason,
		Content:      rep.Message.Content,
		ToolCall:     toolcalls,
	}

	return result, nil
}
