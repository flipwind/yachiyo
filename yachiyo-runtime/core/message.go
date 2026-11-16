package core

import (
	"context"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/yerror"
)

const keepNote = "-1"

func debugOutput(answer []string, c_formal Snapshot, c_later Snapshot) {
	ylog.Debug(`== DEBUG MESSAGE ==
Yachiyo > %v
* Formal:
%s
* Later:
%s
%+v
* Session Context: %s`,
		answer, prompt.Drives(c_formal.State.Drives),
		prompt.Drives(c_later.State.Drives), c_later.State.Determination,
		c_later.Note)
}

func (c *Core) applyOutput(out output) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.State.Apply(out.Change)

	if out.Note != keepNote {
		c.setNote(out.Note)
	}
}

func (c *Core) processUserMessage(m trigger.UserMessage) action.Action {
	rendered := prompt.RenderUserMessage(m)
	ylog.Info("Received user message [%v]", rendered)

	if c.checkTokenThreshold(85, len(rendered)) {
		ylog.Success("Context too big.(85%%) Compressing, please wait...")
		if err := c.processCompress(); err != nil {
			return errorAction(err, m.Address)
		}
	}

	snap := c.snapshot()

	userMessage := event.UserMessage{
		Meta:     c.newMeta(snap.State),
		Content:  m.Content,
		Author:   m.Author,
		Platform: m.Platform,
		Address:  m.Address,
	}

	if err := c.record(userMessage); err != nil {
		ylog.Error("UserMessage history appended error: %s", err)
		return errorAction(yerror.ErrMessageStore, m.Address)
	}

	prompts := prompt.Build(prompt.Input{
		Cause:        prompt.Passive,
		SystemPrompt: c.getSystemPrompt(),
		Window:       c.getWindow(),
		State:        snap.State,
		Note:         c.getNote(),
	})

	tools := GetTools(snap)
	out, err := generate(context.Background(), c.LLM, prompts, tools)

	if err != nil {
		ylog.Error("Process UserMessage failed: %s", err)
		return errorAction(err, m.Address)
	}
	c.applyOutput(out)
	debugOutput(out.Answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()

	assistantMessage := event.AssistantMessage{
		Meta:       c.newMeta(snap.State),
		Content:    out.Answer,
		Address:    m.Address,
		Initiative: false,
		Reply:      out.Reply,
	}

	if err := c.record(assistantMessage); err != nil {
		ylog.Error("Passive InitiativeMessage history appended error: %s", err)
		return errorAction(yerror.ErrMessageStore, m.Address)
	}

	ylog.Success("Generated passive output %#v", out.Answer)

	if c.checkTokenThreshold(70, 0) {
		ylog.Success("Context too big.(70%%) Compressing on background...")
		go c.processCompress()
	}

	return &action.AssistantMessage{
		Address:    assistantMessage.Address,
		Content:    assistantMessage.Content,
		Timestamp:  assistantMessage.Timestamp,
		Initiative: assistantMessage.Initiative,
		Reply:      assistantMessage.Reply,
	}
}

func (c *Core) processInitiativeMessage(_ trigger.InitiativeMessage) action.Action {
	// Check if this trigger vaild.
	snap := c.snapshot()
	window := c.getWindow()
	if len(window.Events) == 0 {
		return nil
	}

	addr, ok := conversation.LastUserAddress(window.Events)
	if ok == false {
		return nil
	}

	ylog.Info("Initiative Message triggered.")

	if c.checkTokenThreshold(85, 0) {
		ylog.Success("Context too big.(85%%) Compressing, please wait...")
		if err := c.processCompress(); err != nil {
			return errorAction(err, addr)
		}
	}

	// Build Prompt
	prompts := prompt.Build(prompt.Input{
		Cause:        prompt.Initiative,
		SystemPrompt: c.getSystemPrompt(),
		Window:       c.getWindow(),
		Note:         c.getNote(),
		State:        snap.State,
	})

	tools := GetTools(snap)
	out, err := generate(context.Background(), c.LLM, prompts, tools)

	if err != nil {
		ylog.Error("Process InitiativeMessage failed: %s", err)
		return errorAction(err, addr)
	}
	c.applyOutput(out)
	debugOutput(out.Answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()

	assistantMessage := event.AssistantMessage{
		Meta:       c.newMeta(snap.State),
		Content:    out.Answer,
		Address:    addr,
		Initiative: true,
		Reply:      out.Reply,
	}

	if err := c.record(assistantMessage); err != nil {
		ylog.Error("Active InitiativeMessage history appended error: %s", err)
		return errorAction(yerror.ErrMessageStore, addr)
	}

	ylog.Success("Generated active output %#v", out.Answer)

	if c.checkTokenThreshold(70, 0) {
		ylog.Success("Context too big.(70%%) Compressing on background...")
		go c.processCompress()
	}
	return &action.AssistantMessage{
		Address:    assistantMessage.Address,
		Content:    assistantMessage.Content,
		Timestamp:  assistantMessage.Timestamp,
		Initiative: assistantMessage.Initiative,
		Reply:      assistantMessage.Reply,
	}
}
