package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/tool"
	"yachiyo/yachiyo-util/yerror"
)

type EmotionPress struct {
	Emotion   string  `json:"emotion"`
	Intensity string  `json:"intensity"`
	Value     float64 `json:"value"`
}

type LLMOutput struct {
	Reply  bool     `json:"reply"`
	Answer []string `json:"answer"`
	Change struct {
		State struct {
			SocialDesire int64 `json:"SocialDesire"`
			Interest     int64 `json:"Interest"`
		} `json:"state_delta"`
		AffectPress []EmotionPress `json:"affect_press"`
	} `json:"change"`
	Determination state.Determination `json:"determination"`
	Note          string              `json:"note"`
}

var emptyAnswers []string = make([]string, 0)

func parseOutput(schema string) (LLMOutput, error) {
	var output LLMOutput

	if err := json.Unmarshal([]byte(schema), &output); err != nil {
		return LLMOutput{}, fmt.Errorf("Json unmarshal error: %w", err)
	}

	if len(output.Answer) == 0 {
		return LLMOutput{}, fmt.Errorf("LLM returns invaild answer: answer has no items")
	}

	if output.Reply == false && len(output.Answer) != 1 {
		return LLMOutput{}, fmt.Errorf("LLM returns invaild answer: choose to not reply but the item number of answer not equals to 1")
	}

	for _, affect := range output.Change.AffectPress {
		if _, ok := state.OCCEmotionMap[affect.Emotion]; !ok {
			return LLMOutput{}, fmt.Errorf("affect emotion type invalid")
		}

		val := affect.Value
		switch affect.Intensity {
		case "slight":
			if val < 0 || val > 1.0/3 {
				return LLMOutput{}, fmt.Errorf("affect emotion value not match intensity")
			}
		case "moderate":
			if val < 1.0/3 || val > 2.0/3 {
				return LLMOutput{}, fmt.Errorf("affect emotion value not match intensity")
			}
		case "high":
			if val < 2.0/3 || val > 1 {
				return LLMOutput{}, fmt.Errorf("affect emotion value not match intensity")
			}
		default:
			return LLMOutput{}, fmt.Errorf("affect emotion intensity invalid")
		}
	}

	ylog.Debug("Parsed LLM output schema: `%s`", schema)

	return output, nil
}

func (c *Core) apply(schema string) ([]string, bool, error) {
	var output LLMOutput
	var err error

	if output, err = parseOutput(schema); err != nil {
		return []string{}, false, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	switch output.Change.State.SocialDesire {
	case 1:
		c.State.SocialDesire.Increase()
	case -1:
		c.State.SocialDesire.Decrease()
	}

	switch output.Change.State.Interest {
	case 1:
		c.State.Interest.Increase()
	case -1:
		c.State.Interest.Decrease()
	}

	// affect
	for _, affect := range output.Change.AffectPress {
		val := affect.Value
		emot := state.OCCEmotionMap[affect.Emotion]
		c.Affect.Emotions = append(c.Affect.Emotions, state.OCCEmotion{
			Type:      emot,
			Intensity: val,
		})
	}

	c.Determination = output.Determination

	if output.Note != "-1" {
		c.Note = output.Note
	}

	allempty := true
	for _, ans := range output.Answer {
		if strings.TrimSpace(ans) != "" {
			allempty = false
			break
		}
	}

	if output.Reply == false || allempty {
		return output.Answer, false, nil
	}
	return output.Answer, true, nil
}

// returns `answers`, `if_reply`, `error`
func (c *Core) processLLM(messages []llm.Message, tools ...tool.Tool) ([]string, bool, error) {
	var jsonConstraint = false
	var lastErr error

	var retryNum = 0

	toolSpec := make([]llm.ToolSpec, 0)
	toolMap := make(map[string]tool.Tool)
	for _, t := range tools {
		toolSpec = append(toolSpec, t.Spec)
		toolMap[t.Spec.Name] = t
	}

	for step := 1; step <= 10; step++ {
		ylog.Info("LLM step %v: requesting model", step)
		if retryNum >= 3 {
			return emptyAnswers, false, fmt.Errorf("LLM failed after 3 attempts: %w", errors.Join(yerror.ErrLLMGenerate, lastErr))
		}

		if jsonConstraint {
			messages = append(messages, llm.Message{
				Role:    llm.RoleSystem,
				Content: "<PROCESS HINT> JSON MODE IS ENABLED. YOU MUST FOLLOW THE OUTPUT ROLE.",
			})
		}

		response, err := c.LLM.Generate(context.Background(), llm.Request{
			Messages: messages,
			JSON:     true,
			Tools:    toolSpec,
		})

		if err != nil {
			lastErr = err
			ylog.Error("LLM Generating error, retrying: %v", err)
			retryNum++
			continue
		}

		switch response.FinishReason {
		case llm.FinishStop:
			answer, reply, err := c.apply(response.Content)
			if err != nil {
				jsonConstraint = true
				lastErr = fmt.Errorf("LLM output unusable: %w", err)
				retryNum++
				continue
			}
			ylog.Info("LLM finished after %v steps", step)
			// If successfully generated, return.
			return answer, reply, nil
		case llm.FinishToolCall:
			messages = append(messages, llm.Message{
				Role:      llm.RoleAssistant,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			})

			for _, call := range response.ToolCalls {
				tool, ok := toolMap[call.Name]
				output := fmt.Sprintf("Error: Unknown tool <`%v`>", call.Name)

				if ok && tool.Execute != nil {
					ylog.Info("LLM using tool call: `%v`(%v), ID [%v]", call.Name, call.Arguments, call.ID)
					res, err := toolMap[call.Name].Execute(call.Arguments)
					if err != nil {
						output = fmt.Sprintf("Error when execute tool <%v>: %v", call.Name, err.Error())
					} else {
						output = res
					}
				}

				ylog.Info("LLM tool result: `%v` -> `%v`, ID [%v]", call.Name, output, call.ID)
				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					Content:    output,
					ToolCallID: call.ID,
				})
			}
		case llm.FinishLength:
			return emptyAnswers, false, fmt.Errorf("Model output exceeded limit: %w", yerror.ErrLLMGenerate)
		case llm.FinishContentFilter:
			return emptyAnswers, false, fmt.Errorf("Output content is being filtered: %w", yerror.ErrLLMGenerate)
		case llm.FinishUnavailable:
			return emptyAnswers, false, fmt.Errorf("Provider now unavailable: %w", yerror.ErrLLMGenerate)
		case llm.FinishAborted:
			return emptyAnswers, false, fmt.Errorf("Output was aborted: %w", yerror.ErrLLMGenerate)
		default:
			lastErr = fmt.Errorf("Unknown finish_reason: %v", response.FinishReason)
			ylog.Error("Unexcepted finish_reason, retrying: %v", response.FinishReason)
			retryNum++
			continue
		}
	}

	return emptyAnswers, false, fmt.Errorf("LLM exceeded 10 steps, stop generating: %w", errors.Join(yerror.ErrLLMGenerate, lastErr))
}
