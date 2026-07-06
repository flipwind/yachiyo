package core

import (
	"encoding/json"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Process(e *trigger.Message) string {
	histories := prompt.PromptBuilder(&prompt.Context{
		History: c.History,
		Emotion: c.Emotion,
		State:   c.State,
		Note:    c.Note,
	}, e)

	var result, answer string
	var err error

	for i := range 3 {
		if i == 1 {
			c.JSONConstraint = true
		}

		if c.JSONConstraint {
			histories = append(histories, history.History{
				Role:    "user",
				Content: "<PROCESS HINT> YOUR LAST REPLY IS NOT A VALID JSON, REGENERATE IT. YOU MUST FOLLOW THE OUTPUT ROLE.",
			})
		}

		result, err = c.LLM.Gen(histories)

		if err != nil {
			ylog.Error("LLM Generating error: %v", err)
			continue
		}

		answer, err = c.OutputProcess(result)

		if err != nil {
			ylog.Error("%d request failed. Retrying...", i+1)
			continue
		}

		c.History.Remember(history.History{
			Role:    "assistant",
			Content: answer,
		})

		break
	}

	return answer
}

type LLMOutput struct {
	Answer string `json:"answer"`
	Change struct {
		Emotion struct {
			Change string `json:"emotion"`
			Result string `json:"result"`
		} `json:"emotion"`
		State struct {
			SocialDesire int64 `json:"SocialDesire"`
			Interest      int64 `json:"Interest"`
		} `json:"state_delta"`
	} `json:"change"`
	Determination state.Determination `json:"determination"`
	Note string `json:"note"`
}

func (c *Core) OutputProcess(schema string) (string, error) {
	var output LLMOutput
	if err := json.Unmarshal([]byte(schema), &output); err != nil {
		ylog.Error("Json unmarshal error: %v", err)
		return "", err
	}

	c.Emotion.Type = output.Change.Emotion.Change
	c.Emotion.Urgency = state.UrgencyFromString(output.Change.Emotion.Result)

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

	c.Determination = output.Determination

	if output.Note != "-1" {
		c.Note = output.Note
	}

	return output.Answer, nil
}
