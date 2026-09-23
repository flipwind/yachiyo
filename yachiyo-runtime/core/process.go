package core

import (
	"encoding/json"
	"fmt"
	"strings"
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

func (c *Core) apply(schema string) (string, bool, error) {
	var output LLMOutput
	if err := json.Unmarshal([]byte(schema), &output); err != nil {
		ylog.Error("Json unmarshal error: %v", err)
		return "", false, err
	}

	ylog.Debug("%s", schema)

	c.mu.Lock()
	defer c.mu.Unlock()

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

	if output.Reply == false || strings.TrimSpace(output.Answer) == "" {
		return "Yachiyo didn't reply.", false, nil
	}
	return output.Answer, true, nil
}

// utils
func debugOutput(answer string, c_formal Snapshot, c_later Snapshot) {
	ylog.Debug(`== DEBUG MESSAGE ==
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
}

func (c *Core) processLLM(prompts []prompt.Prompts) (string, bool) {
	var answer string
	var reply bool

	for i := range 3 {
		if i == 1 {
			// This is desiged intentionally.
			// In real use, once triggered, JSONConstraint should be true in a whole session,
			// to cut down the token use.
			c.mu.Lock()
			c.JSONConstraint = true
			c.mu.Unlock()
		}

		c.mu.Lock()
		isJSONConstraint := c.JSONConstraint
		c.mu.Unlock()

		if isJSONConstraint {
			prompts = append(prompts, prompt.SystemPrompt{
				Content: "<PROCESS HINT> JSON MODE IS ENABLED. YOU MUST FOLLOW THE OUTPUT ROLE.",
			})
		}

		result, err := c.LLM.Gen(prompts)

		if err != nil {
			ylog.Error("LLM Generating error: %v", err)
			continue
		}

		answer, reply, err = c.apply(result)

		if err != nil {
			ylog.Error("%d request failed. Retrying...", i+1)
			continue
		}

		if reply == false {
			break
		}

		break
	}

	return answer, reply
}

// Trigger process part

func (c *Core) processUserMessage(m *trigger.Message) action.Action {
	snap := c.snapshot()
	histSnap := history.Snapshot{
		Time:           time.Now(),
		Emotion:        snap.Emotion,
		State:          snap.State,
		Factors:        snap.Factors,
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
	}, m, histSnap)

	answer, isReply := c.processLLM(prompts)
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = history.Snapshot{
		Time:           time.Now(),
		Emotion:        snap.Emotion,
		State:          snap.State,
		Factors:        snap.Factors,
		LastActiveTime: snap.LastActiveTime,
	}

	c.AppendHistory(history.AssistantMessage{
		Snapshot: histSnap,
		Content:  answer,

		Time:      time.Now(),
		ToAddress: m.Address,
		Note:      c.getNote(),

		IsInitiative: false,
		IsEmpty:      !isReply,
	})

	ylog.Success("Generated passive output [%v]", answer)
	return &action.Message{
		Empty:   !isReply,
		Content: answer,
		Time:    time.Now().Unix(),
		Address: m.Address,
	}
}

func (c *Core) processInitiativeMessage(_ *trigger.InitiativeMessage) action.Action {
	// Check if this trigger vaild.
	snap := c.snapshot()
	hist := c.getHistory()
	if len(hist) == 0 {
		return nil
	}

	addr, ok := GetLastUserAddress(hist)
	if ok == false {
		return nil
	}

	histSnap := history.Snapshot{
		Time:           time.Now(),
		Emotion:        snap.Emotion,
		State:          snap.State,
		Factors:        snap.Factors,
		LastActiveTime: snap.LastActiveTime,
	}

	// Build Prompt
	prompts := prompt.InitiativePromptBuilder(
		prompt.Context{
			SystemPrompt: c.getSystemPrompt(),
			History:      c.getHistory(),
			Note:         c.getNote(),
		}, histSnap)

	answer, isReply := c.processLLM(prompts)
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = history.Snapshot{
		Time:           time.Now(),
		Emotion:        snap.Emotion,
		State:          snap.State,
		Factors:        snap.Factors,
		LastActiveTime: snap.LastActiveTime,
	}

	c.AppendHistory(history.AssistantMessage{
		Snapshot: histSnap,
		Content:  answer,

		Time:      time.Now(),
		ToAddress: addr,
		Note:      c.getNote(),

		IsInitiative: true,
		IsEmpty:      !isReply,
	})

	ylog.Success("Generated active output [%v]", answer)
	return &action.Message{
		Empty:   !isReply,
		Content: answer,
		Time:    time.Now().Unix(),
		Address: addr,
	}
}

func (c *Core) processRuntimeStateRequest(t *trigger.RuntimeStateRequest) action.Action {
	DebugMessage := fmt.Sprintf("Received timetick %s\n", time.Now().Format("15:04:05"))

	c.mu.Lock()
	for _, s := range c.State.Drives() {
		DebugMessage += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	DebugMessage += fmt.Sprintf("factors: %v\n", c.Factors.String())
	c.mu.Unlock()

	return &action.RuntimeState{
		Content: DebugMessage,
		Address: t.Address,
	}
}
