package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/state"
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

func (c *Core) apply(schema string) ([]string, bool, error) {
	var output LLMOutput
	if err := json.Unmarshal([]byte(schema), &output); err != nil {
		return emptyAnswers, false, fmt.Errorf("Json unmarshal error: %w", err)
	}

	if len(output.Answer) == 0 {
		return emptyAnswers, false, fmt.Errorf("LLM returns invaild answer: answer has no items")
	}

	if output.Reply == false && len(output.Answer) != 1 {
		return emptyAnswers, false, fmt.Errorf("LLM returns invaild answer: choose to not reply but the item number of answer not equals to 1")
	}

	for _, affect := range output.Change.AffectPress {
		if _, ok := state.OCCEmotionMap[affect.Emotion]; !ok {
			return emptyAnswers, false, fmt.Errorf("affect emotion type invalid")
		}

		val := affect.Value
		switch affect.Intensity {
		case "slight":
			if val < 0 || val > 1.0/3 {
				return emptyAnswers, false, fmt.Errorf("affect emotion value not match intensity")
			}
		case "moderate":
			if val < 1.0/3 || val > 2.0/3 {
				return emptyAnswers, false, fmt.Errorf("affect emotion value not match intensity")
			}
		case "high":
			if val < 2.0/3 || val > 1 {
				return emptyAnswers, false, fmt.Errorf("affect emotion value not match intensity")
			}
		default:
			return emptyAnswers, false, fmt.Errorf("affect emotion intensity invalid")
		}
	}

	ylog.Debug("%s", schema)

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

func (c *Core) processLLM(rawPrompts []prompt.Prompts) ([]string, bool, error) {
	var prompts = rawPrompts
	var jsonConstraint = false
	var lastErr error

	for range 3 {
		if jsonConstraint {
			prompts = append(rawPrompts, prompt.SystemPrompt{
				Content: "<PROCESS HINT> JSON MODE IS ENABLED. YOU MUST FOLLOW THE OUTPUT ROLE.",
			})
		}

		result, err, finish_reason := c.LLM.Gen(prompts, true)

		if err != nil {
			lastErr = err
			ylog.Error("LLM Generating error, retrying: %v", err)
			continue
		}

		if finish_reason != "stop" {
			lastErr = fmt.Errorf("finish_reason of LLM generation is not `stop`.")
			ylog.Error("LLM Generating error, retrying: %v", err)
			continue
		}

		answer, reply, err := c.apply(result)

		if err != nil {
			jsonConstraint = true
			lastErr = fmt.Errorf("LLM output unusable: %w", err)
			continue
		}

		// If successfully generated, return.
		return answer, reply, nil
	}

	return emptyAnswers, false, fmt.Errorf("LLM failed after 3 attempts: %w", errors.Join(yerror.ErrLLMGenerate, lastErr))
}
