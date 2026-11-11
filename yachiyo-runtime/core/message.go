package core

import (
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/yerror"
)

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

func (c *Core) processUserMessage(m trigger.UserMessage) action.Action {
	rendered := prompt.RenderUserMessage(m)
	ylog.Info("Received user message [%v]", rendered)

	if c.checkTokenThreshold(85, len(rendered)) {
		ylog.Success("Context too big.(85%%) Compressing, please wait...")
		if err := c.processCompress(); err != nil {
			return processError(err, m.Address)
		}
	}

	snap := c.snapshot()
	histSnap := snap.toHistorySnapshot(time.Now())

	userMessage := event.UserMessage{
		ID:       uuid.NewV7(),
		Sequence: c.History.NextSequence(),
		Snapshot: histSnap,
		Content:  m.Content,

		Author:    m.Author,
		Platform:  m.Platform,
		Timestamp: time.Now(),
		Address:   m.Address,
	}

	if err := c.AppendHistory(userMessage); err != nil {
		ylog.Error("UserMessage history appended error: %s", err)
		return processError(yerror.ErrMessageStore, m.Address)
	}

	prompts := prompt.UserPromptBuilder(prompt.Context{
		SystemPrompt:  c.getSystemPrompt(),
		ContextWindow: c.getContextWindow(),
		Note:          c.getNote(),
	}, histSnap)

	tools := c.GetTools()
	answer, isReply, err := c.processLLM(prompts, tools...)

	if err != nil {
		ylog.Error("Process UserMessage failed: %s", err)
		return processError(err, m.Address)
	}
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = snap.toHistorySnapshot(time.Now())

	assistantMessage := event.AssistantMessage{
		ID:       uuid.NewV7(),
		Sequence: c.History.NextSequence(),
		Snapshot: histSnap,
		Content:  answer,

		Timestamp: time.Now(),
		Address:   m.Address,

		Initiative: false,
		Reply:      isReply,
	}

	if err := c.AppendHistory(assistantMessage); err != nil {
		ylog.Error("Passive InitiativeMessage history appended error: %s", err)
		return processError(yerror.ErrMessageStore, m.Address)
	}

	ylog.Success("Generated passive output %#v", answer)

	if c.checkTokenThreshold(70, 0) {
		ylog.Success("Context too big.(70%%) Compressing on background...")
		go c.processCompress()
	}

	return &action.AssistantMessage{AssistantMessageContent: assistantMessage.AssistantMessageContent}
}

func (c *Core) processInitiativeMessage(_ trigger.InitiativeMessage) action.Action {
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

	if c.checkTokenThreshold(85, 0) {
		ylog.Success("Context too big.(85%%) Compressing, please wait...")
		if err := c.processCompress(); err != nil {
			return processError(err, addr)
		}
	}

	histSnap := snap.toHistorySnapshot(time.Now())

	// Build Prompt
	prompts := prompt.InitiativePromptBuilder(
		prompt.Context{
			SystemPrompt:  c.getSystemPrompt(),
			ContextWindow: c.getContextWindow(),
			Note:          c.getNote(),
		}, histSnap)

	tools := c.GetBuiltinTools()
	answer, isReply, err := c.processLLM(prompts, tools...)

	if err != nil {
		ylog.Error("Process InitiativeMessage failed: %s", err)
		return processError(err, addr)
	}
	debugOutput(answer, snap, c.snapshot())

	// After processed, store the assistant message

	snap = c.snapshot()
	histSnap = snap.toHistorySnapshot(time.Now())

	assistantMessage := event.AssistantMessage{
		ID:       uuid.NewV7(),
		Sequence: c.History.NextSequence(),
		Snapshot: histSnap,
		Content:  answer,

		Timestamp: time.Now(),
		Address:   addr,

		Initiative: true,
		Reply:      isReply,
	}

	if err := c.AppendHistory(assistantMessage); err != nil {
		ylog.Error("Active InitiativeMessage history appended error: %s", err)
		return processError(yerror.ErrMessageStore, addr)
	}

	ylog.Success("Generated active output %#v", answer)

	if c.checkTokenThreshold(70, 0) {
		ylog.Success("Context too big.(70%%) Compressing on background...")
		go c.processCompress()
	}

	return &action.AssistantMessage{AssistantMessageContent: assistantMessage.AssistantMessageContent}
}
