package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/tool"
	"yachiyo/yachiyo-util/yerror"
)

const maxRetries = 3
const maxSteps = 10

type llmOutput struct {
	Thoughts string `json:"thoughts"`
	Change   struct {
		State struct {
			SocialDesire int `json:"SocialDesire"`
			Interest     int `json:"Interest"`
		} `json:"state_delta"`
		AffectPress []struct {
			Emotion   string  `json:"emotion"`
			Intensity string  `json:"intensity"`
			Value     float64 `json:"value"`
		} `json:"affect_press"`
	} `json:"change"`
	Determination struct {
		NewSession   bool `json:"should_open_a_new_session"`
		WaitForReply bool `json:"should_wait_for_reply"`
	} `json:"determination"`
	Note string `json:"note"`
}

type output struct {
	Thoughts string
	Change   state.Change
	Note     string
}

func parseOutput(schema string) (output, error) {
	var llmoutput llmOutput

	if err := json.Unmarshal([]byte(schema), &llmoutput); err != nil {
		return output{}, fmt.Errorf("json unmarshal error: %w", err)
	}

	emotions := make([]state.OCCEmotion, 0)
	for _, affect := range llmoutput.Change.AffectPress {
		typ, ok := state.OCCEmotionMap[affect.Emotion]
		if !ok {
			return output{}, fmt.Errorf("affect emotion type invalid")
		}

		val := affect.Value
		var inf, sup float64
		switch affect.Intensity {
		case "slight":
			inf, sup = 0, 1.0/3
		case "moderate":
			inf, sup = 1.0/3, 2.0/3
		case "high":
			inf, sup = 2.0/3, 1
		default:
			return output{}, errors.New("affect emotion intensity invalid")
		}

		if val < inf || val > sup {
			return output{},
				fmt.Errorf("affect emotion value not match intensity: <%v> should be in [%v, %v], (%v)",
					val, inf, sup, affect.Intensity)
		}

		emotions = append(emotions, state.OCCEmotion{
			Type:      typ,
			Intensity: val,
		})
	}

	ylog.Debug("Parsed LLM output schema: `%s`", schema)

	return output{
		Thoughts: llmoutput.Thoughts,
		Change: state.Change{
			SocialDesire: llmoutput.Change.State.SocialDesire,
			Interest:     llmoutput.Change.State.Interest,
			Emotions:     emotions,
			Determination: state.Determination{
				NewSession:   llmoutput.Determination.NewSession,
				WaitForReply: llmoutput.Determination.WaitForReply,
			},
		},
		Note: llmoutput.Note,
	}, nil
}

func generate(ctx context.Context, client llm.Client, messages []llm.Message, tools tool.Set) (output, error) {
	var lastErr error

	var retryNum = 0

	for step := 1; step <= maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return output{}, err
		}

		ylog.Info("LLM step %v: requesting model", step)
		if retryNum >= maxRetries {
			return output{}, fmt.Errorf("LLM failed after %v attempts: %w", maxRetries, errors.Join(yerror.ErrLLMGenerate, lastErr))
		}

		response, err := client.Generate(ctx, llm.Request{
			Messages: messages,
			JSON:     true,
			Tools:    tools.Specs(),
		})

		if err != nil {
			lastErr = err
			ylog.Error("LLM Generating error, retrying: %v", err)
			retryNum++
			continue
		}

		switch response.FinishReason {
		case llm.FinishStop:
			out, err := parseOutput(response.Content)
			if err != nil {
				messages = append(messages, prompt.JSONReminder())

				lastErr = fmt.Errorf("LLM output unusable: %w", err)
				retryNum++
				continue
			}
			ylog.Info("LLM finished after %v steps", step)
			// If successfully generated, return.
			return out, nil
		case llm.FinishToolCall:
			messages = append(messages, llm.Message{
				Role:      llm.RoleAssistant,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			})

			for _, call := range response.ToolCalls {
				ylog.Info("LLM using tool call: `%v`(%v), ID [%v]", call.Name, call.Arguments, call.ID)
				out := tools.Call(call.Name, call.Arguments)

				ylog.Info("LLM tool result: `%v` -> `%v`, ID [%v]", call.Name, out, call.ID)
				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					Content:    out,
					ToolCallID: call.ID,
				})
			}
		case llm.FinishLength:
			return output{}, fmt.Errorf("model output exceeded limit: %w", yerror.ErrLLMGenerate)
		case llm.FinishContentFilter:
			return output{}, fmt.Errorf("output content is being filtered: %w", yerror.ErrLLMGenerate)
		case llm.FinishUnavailable:
			return output{}, fmt.Errorf("provider now unavailable: %w", yerror.ErrLLMGenerate)
		case llm.FinishAborted:
			return output{}, fmt.Errorf("output was aborted: %w", yerror.ErrLLMGenerate)
		default:
			lastErr = fmt.Errorf("unknown finish_reason: %v", response.FinishReason)
			ylog.Error("Unexcepted finish_reason, retrying: %v", response.FinishReason)
			retryNum++
			continue
		}
	}

	return output{}, fmt.Errorf("LLM exceeded %v steps, stop generating: %w", maxSteps, errors.Join(yerror.ErrLLMGenerate, lastErr))
}
