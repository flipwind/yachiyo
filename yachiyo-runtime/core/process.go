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
	switch t := e.(type) {
	case *trigger.Message:
		return c.processUserMessage(t)
	case *trigger.TimeTick:
		c.processTimetick(t)
	case *trigger.InitiativeMessage:
		return c.processInitiativeMessage(t)
	default:
		ylog.Error("Process unsupported type: %T", t)
		return nil
	}
	return nil
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

// Trigger process part

func (c *Core) processUserMessage(m *trigger.Message) action.Action {
	histories := prompt.UserPromptBuilder(&prompt.Context{
		History: c.History,
		Emotion: c.Emotion,
		State:   c.State,
		Note:    c.Note,
		Factors: c.Factors,
	}, m)

	var result, answer string
	var err error

	// <debug>
	var core_copy Core
	core_copy = *c
	// </debug>

	for i := range 3 {
		if i == 1 {
			c.JSONConstraint = true
		}

		if c.JSONConstraint {
			histories = append(histories, history.History{
				Role:    "user",
				Content: "<PROCESS HINT> YOUR LAST REPLY IS NOT A VALID JSON, REGENERATE IT. YOU MUST FOLLOW THE OUTPUT ROLE.",
				Time:    time.Now(),
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
			Time:    time.Now(),
		})

		break
	}

	// <debug>
	fmt.Printf("Yachiyo > %s\n", answer)

	fmt.Printf("\nDEBUG\nFORMAL:")
	fmt.Println(core_copy.Emotion.String())
	fmt.Println(core_copy.State.Prompt())
	fmt.Printf("\nLATER:\n")
	fmt.Println(c.Emotion.String())
	fmt.Println(c.State.Prompt())
	fmt.Println(c.Determination)

	fmt.Println("Session Context: " + c.Note)
	// </debug>

	return &action.Message{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: action.Address{
			Content: m.Address.Content,
		},
	}
}

func (c *Core) processInitiativeMessage(_ *trigger.InitiativeMessage) action.Action {
	histories := prompt.InitiativePromptBuilder(&prompt.Context{
		History: c.History,
		Emotion: c.Emotion,
		State:   c.State,
		Note:    c.Note,
		Factors: c.Factors,
	})

	var result, answer string
	var err error

	// <debug>
	var core_copy Core
	core_copy = *c
	// </debug>

	for i := range 3 {
		if i == 1 {
			c.JSONConstraint = true
		}

		if c.JSONConstraint {
			histories = append(histories, history.History{
				Role:    "user",
				Content: "<PROCESS HINT> YOUR LAST REPLY IS NOT A VALID JSON, REGENERATE IT. YOU MUST FOLLOW THE OUTPUT ROLE.",
				Time:    time.Now(),
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
			Time:    time.Now(),
		})

		break
	}

	// <debug>
	fmt.Printf("Yachiyo > %s\n", answer)

	fmt.Printf("\nDEBUG\nFORMAL:")
	fmt.Println(core_copy.Emotion.String())
	fmt.Println(core_copy.State.Prompt())
	fmt.Printf("\nLATER:\n")
	fmt.Println(c.Emotion.String())
	fmt.Println(c.State.Prompt())
	fmt.Println(c.Determination)

	fmt.Println("Session Context: " + c.Note)
	// </debug>

	return &action.Message{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: action.Address{
			Content: "client://cli",
		},
	}
}

func (c *Core) processTimetick(t *trigger.TimeTick) {
	ylog.Debug("Received timetick %s", t.Time.Format("15:04:05"))

	// State press
	for _, s := range c.State.Drives() {
		s.Drive.Press(0.1) // TODO: 0.1 is high speed, being only in debug
		ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
	}

	// Initiative
	timeNow := time.Now()
	alonetime := time.Since(c.History.GetLastActive())
	daytime := timeNow.Sub(time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()))
	ylog.Debug("factors: %v", c.Factors.String())
	if c.Factors.Update(alonetime.Minutes(), daytime.Minutes()) {
		ylog.Info("Initiative active.")

		// Relieve
		for _, s := range c.State.Drives() {
			s.Drive.Relieve(0.5) // TODO: 0.5 is high speed, being only in debug
			ylog.Debug("State: %v at %v, is %v", s.Name, s.Drive.Value, s.Drive.String())
		}

		c.Process(&trigger.InitiativeMessage{})
	}
}
