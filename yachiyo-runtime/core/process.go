package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/yerror"
)

func (c *Core) Process(e trigger.Trigger) action.Action {
	timeNow := time.Now()

	switch t := e.(type) {
	case *trigger.Message:
		c.mu.Lock()
		c.LastActiveTime = timeNow

		c.LLMBusy = true
		c.mu.Unlock()
		a := c.processUserMessage(t)

		c.mu.Lock()
		c.LLMBusy = false
		c.mu.Unlock()

		return a
	case *trigger.InitiativeMessage:
		c.mu.Lock()
		c.LastActiveTime = timeNow

		c.LLMBusy = true
		c.mu.Unlock()
		a := c.processInitiativeMessage(t)

		c.mu.Lock()
		c.LLMBusy = false
		c.mu.Unlock()

		return a
	case *trigger.RuntimeStateRequest:
		return c.processRuntimeStateRequest(t)
	case *trigger.MessageHistoryRequest:
		return c.processMessageHistoryRequest(t)
	default:
		ylog.Error("Process unsupported type: %T", t)
		return nil
	}
}

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

// utils
func debugOutput(answer []string, c_formal Snapshot, c_later Snapshot) {
	ylog.Debug(`== DEBUG MESSAGE ==
Yachiyo > %v
* Formal:
%s
* Later:
%s
%+v
* Session Context: %s`,
		answer, c_formal.State.Prompt(),
		c_later.State.Prompt(), c_later.Determination,
		c_later.Note)
}

func processError(err error, addr address.Address) action.Action {
	if errors.Is(err, yerror.ErrLLMChoice) {
		return &action.Error{
			Code:    action.ErrLLMChoiceEmpty,
			Message: err.Error(),
			Address: addr,
		}
	}

	if errors.Is(err, yerror.ErrLLMGenerate) {
		return &action.Error{
			Code:    action.ErrLLMGenerate,
			Message: err.Error(),
			Address: addr,
		}
	}

	return &action.Error{
		Code:    action.ErrMessageProcess,
		Message: err.Error(),
		Address: addr,
	}
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

		result, err := c.LLM.Gen(prompts)

		if err != nil {
			lastErr = err
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

// Trigger process part

func (c *Core) processUserMessage(m *trigger.Message) action.Action {
	ylog.Info("Received user message [%v]", m.String())
	snap := c.snapshot()
	histSnap := history.Snapshot{
		Time:           time.Now(),
		Affect:         snap.Affect,
		State:          snap.State,
		Energy:         snap.Energy,
		LastActiveTime: snap.LastActiveTime,
	}

	c.AppendHistory(history.UserMessage{
		Snapshot: histSnap,
		Content:  m.Content,

		Author:   m.Author,
		Platform: m.Platform,
		Time:     time.Now(),
		Address:  m.Address,
	})

	prompts := prompt.UserPromptBuilder(prompt.Context{
		SystemPrompt: c.getSystemPrompt(),
		History:      c.getHistory(),
		Note:         c.getNote(),
	}, histSnap)

	answer, isReply, err := c.processLLM(prompts)
	if err != nil {
		ylog.Error("Process UserMessage failed: %s", err)
		return processError(err, m.Address)
	}
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = history.Snapshot{
		Time:           time.Now(),
		Affect:         snap.Affect,
		State:          snap.State,
		Energy:         snap.Energy,
		LastActiveTime: snap.LastActiveTime,
	}

	c.AppendHistory(history.AssistantMessage{
		Snapshot: histSnap,
		Content:  answer,

		Time:      time.Now(),
		ToAddress: m.Address,
		Note:      c.getNote(),

		Initiative: false,
		Reply:      isReply,
	})

	ylog.Success("Generated passive output %#v", answer)
	return &action.AssistantMessage{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: m.Address,

		Initiative: false,
		Reply:      isReply,
	}
}

func (c *Core) processInitiativeMessage(_ *trigger.InitiativeMessage) action.Action {
	// Check if this trigger vaild.
	snap := c.snapshot()
	hist := c.getHistory()
	if len(hist) == 0 {
		return nil
	}

	addr, ok := history.GetLastUserAddress(hist)
	if ok == false {
		return nil
	}

	ylog.Info("Initiative Message triggered.")

	histSnap := history.Snapshot{
		Time:           time.Now(),
		Affect:         snap.Affect,
		State:          snap.State,
		Energy:         snap.Energy,
		LastActiveTime: snap.LastActiveTime,
	}

	// Build Prompt
	prompts := prompt.InitiativePromptBuilder(
		prompt.Context{
			SystemPrompt: c.getSystemPrompt(),
			History:      c.getHistory(),
			Note:         c.getNote(),
		}, histSnap)

	answer, isReply, err := c.processLLM(prompts)
	if err != nil {
		ylog.Error("Process InitiativeMessage failed: %s", err)
		return processError(err, addr)
	}
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = history.Snapshot{
		Time:           time.Now(),
		Affect:         snap.Affect,
		State:          snap.State,
		Energy:         snap.Energy,
		LastActiveTime: snap.LastActiveTime,
	}

	c.AppendHistory(history.AssistantMessage{
		Snapshot: histSnap,
		Content:  answer,

		Time:      time.Now(),
		ToAddress: addr,
		Note:      c.getNote(),

		Initiative: true,
		Reply:      isReply,
	})

	ylog.Success("Generated active output %#v", answer)
	return &action.AssistantMessage{
		Content: answer,
		Time:    time.Now().Unix(),
		Address: addr,

		Initiative: true,
		Reply:      isReply,
	}
}

func (c *Core) processRuntimeStateRequest(t *trigger.RuntimeStateRequest) action.Action {
	actives := make([]action.OCCEmotion, 0)
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, emo := range c.Affect.Emotions {
		actives = append(actives, action.OCCEmotion{
			Type:      emo.Type.Name,
			Intensity: emo.Intensity,
		})
	}

	return &action.RuntimeState{
		Timetick: time.Now().Unix(),
		Energy:   c.Energy.Value,
		Affect: struct {
			Octants   string
			Pleasure  float64
			Arousal   float64
			Dominance float64
			Actives   []action.OCCEmotion
		}{
			Octants:   c.Affect.Current.Octants(),
			Pleasure:  c.Affect.Current.Pleasure(),
			Arousal:   c.Affect.Current.Arousal(),
			Dominance: c.Affect.Current.Dominance(),
			Actives:   actives,
		},
		Address: t.Address,
	}
}

func (c *Core) processMessageHistoryRequest(t *trigger.MessageHistoryRequest) action.Action {
	hist := c.getHistory()
	msgs := make([]action.Message, 0)

	for _, h := range hist {
		switch m := h.(type) {
		case history.AssistantMessage:
			if m.ToAddress != t.Address {
				break
			}

			if m.Initiative == true && m.Reply == false {
				break
			}

			msgs = append(msgs, &action.AssistantMessage{
				Content:    m.Content,
				Time:       m.Time.Unix(),
				Initiative: m.Initiative,
				Reply:      m.Reply,
				Address:    m.ToAddress,
			})
		case history.UserMessage:
			if m.Address != t.Address {
				break
			}

			msgs = append(msgs, &action.UserMessage{
				Content: m.Content,
				Time:    m.Time.Unix(),
				Address: m.Address,
			})
		}
	}

	return &action.MessageHistory{
		Messages: msgs,
		Address:  t.Address,
	}
}
