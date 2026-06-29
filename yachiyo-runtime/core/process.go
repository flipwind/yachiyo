package core

import (
	"encoding/json"
	"yachiyo/yachiyo-runtime/state"
)

type LLMOutput struct {
	Answer string `json:"answer"`
	Change struct {
		Emotion struct {
			Change string `json:"emotion"`
			Result string `json:"result"`
		} `json:"emotion"`
		State struct {
			Socialization int64 `json:"Socialization"`
			Interest      int64 `json:"Interest"`
		} `json:"state_delta"`
	} `json:"change"`
}

func (c *Core) OutputProcess(schema string) (string, error) {
	var output LLMOutput
	if err := json.Unmarshal([]byte(schema), &output); err != nil {
		ylog.Error("Json unmarshal error: %v", err)
		return "", err
	}

	c.Emotion.Type = output.Change.Emotion.Change
	c.Emotion.Urgency = state.UrgencyFromString(output.Change.Emotion.Result)

	switch output.Change.State.Socialization {
	case 1:
		c.State.Socialization.Increase()
	case -1:
		c.State.Socialization.Decrease()
	}

	switch output.Change.State.Interest {
	case 1:
		c.State.Interest.Increase()
	case -1:
		c.State.Interest.Decrease()
	}

	return output.Answer, nil
}
