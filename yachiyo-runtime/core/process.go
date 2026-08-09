package core

import (
	"encoding/json"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) Process(e trigger.Trigger) action.Action {
	timeNow := time.Now()

	switch t := e.(type) {
	case *trigger.Message:
		c.LastActiveTime = &timeNow

		c.LLMBusy = true
		a := c.processUserMessage(t)
		c.LLMBusy = false

		return a
	case *trigger.InitiativeMessage:
		c.LastActiveTime = &timeNow

		c.LLMBusy = true
		a := c.processInitiativeMessage(t)
		c.LLMBusy = false

		return a
	default:
		ylog.Error("Process unsupported type: %T", t)
		return nil
	}
}

type LLMOutput struct {
	Reply  bool   `json:"reply"`
	Answer string `json:"answer"`
	Change struct {
		Emotion struct {
			Change string `json:"emotion"`
			Result string `json:"result"`
		} `json:"emotion"`
		State struct {
			SocialDesire int64 `json:"SocialDesire"`
			Interest     int64 `json:"Interest"`
		} `json:"state_delta"`
	} `json:"change"`
	Determination state.Determination `json:"determination"`
	Note          string              `json:"note"`
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

	if output.Reply == false {
		return "Yachiyo didn't reply.", nil
	}
	return output.Answer, nil
}

// utils
func debugOutput(answer string, c_formal Core, c_later Core) {
	msg := fmt.Sprintf(`== DEBUG MESSAGE ==
Yachiyo > %s
* Formal:
%s
%s
* Later:
%s
%s
%+v
* Session Context: %s`,
		answer, c_formal.Emotion.String(), c_formal.State.Prompt(),
		c_later.Emotion.String(), c_later.State.Prompt(), c_later.Determination,
		c_later.Note)

	fmt.Printf("%s", msg)
}

func processLLM(h []history.History, c *Core) string {
	var answer string

	for i := range 3 {
		if i == 1 {
			// This is desiged intentionally.
			// In real use, once triggered, JSONConstraint should be true in a whole session,
			// to cut down the token use.
			c.JSONConstraint = true
		}

		if c.JSONConstraint {
			h = append(h, history.History{
				Role:    "user",
				Content: "<PROCESS HINT> JSON MODE IS ENABLED. YOU MUST FOLLOW THE OUTPUT ROLE.",
				Time:    time.Now(),
			})
		}

		result, err := c.LLM.Gen(h)

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
			Time:    time.Now(),
		})

		break
	}

	return answer
}

// Trigger process part

func (c *Core) processUserMessage(m *trigger.Message) action.Action {
	histories := prompt.UserPromptBuilder(&prompt.Context{
		SystemPrompt:   c.Config.Prompt.SystemPrompt,
		History:        c.History,
		Emotion:        c.Emotion,
		State:          c.State,
		Note:           c.Note,
		Factors:        c.Factors,
		LastActiveTime: *c.LastActiveTime,
	}, m)

	// <debug>
	var core_copy Core
	core_copy = *c
	// </debug>

	answer := processLLM(histories, c)

	debugOutput(answer, core_copy, *c)

	return &action.Message{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: m.Address,
	}
}

func (c *Core) processInitiativeMessage(_ *trigger.InitiativeMessage) action.Action {
	histories := prompt.InitiativePromptBuilder(&prompt.Context{
		SystemPrompt:   c.Config.Prompt.SystemPrompt,
		History:        c.History,
		Emotion:        c.Emotion,
		State:          c.State,
		Note:           c.Note,
		Factors:        c.Factors,
		LastActiveTime: *c.LastActiveTime,
	})

	// <debug>
	var core_copy Core
	core_copy = *c
	// </debug>

	answer := processLLM(histories, c)

	debugOutput(answer, core_copy, *c)

	return &action.Message{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: c.History.GetLastUserHistory().Address,
	}
}


