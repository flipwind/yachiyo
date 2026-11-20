package core

import (
	"errors"
	"time"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/yerror"
)

const keepNote = "-1"

type userJob struct {
	msg        trigger.UserMessage
	compressed bool
}

func (userJob) isJob() {}

type initiativeJob struct {
	compressed bool
}

func (initiativeJob) isJob() {}

type reply struct {
	cause prompt.Cause
	addr  address.Address
	snap  Snapshot
	out   output
	err   error
}

func (reply) isResult() {}

func (c *Core) startUser(j userJob) {
	rendered := prompt.RenderUserMessage(j.msg)

	if !j.compressed && c.checkTokenThreshold(85, len(rendered)) {
		ylog.Warn("Context too big.(85%%) Compressing, please wait...")
		j.compressed = true
		c.queue = append([]job{compressJob{}, j}, c.queue...)
		return
	}

	ylog.Info("Now processing user message [%v]", rendered)

	snap := c.snapshot()

	userMessage := event.UserMessage{
		Meta:     c.newMeta(snap.State),
		Content:  j.msg.Content,
		Author:   j.msg.Author,
		Platform: j.msg.Platform,
		Address:  j.msg.Address,
	}

	if err := c.record(userMessage); err != nil {
		ylog.Error("UserMessage history appended error: %s", err)
		c.deliver(errorAction(yerror.ErrMessageStore, j.msg.Address))
		return
	}

	c.startGenerate(prompt.Passive, j.msg.Address)
}

func (c *Core) startInitiative(j initiativeJob) {
	if !j.compressed && c.checkTokenThreshold(85, 0) {
		ylog.Success("Context too big.(85%%) Compressing, please wait...")
		j.compressed = true
		c.queue = append([]job{compressJob{}, j}, c.queue...)
		return
	}

	ylog.Info("Initiative Message triggered.")

	// Check if this trigger vaild.
	window := c.conv.Window()
	if len(window.Events) == 0 {
		ylog.Warn("Not histories found, skip this turn")
		return
	}

	addr, ok := conversation.LastUserAddress(window.Events)
	if ok == false {
		ylog.Warn("Can't find last user address, ignore this turn")
		return
	}

	c.startGenerate(prompt.Initiative, addr)
}

func (c *Core) startGenerate(cause prompt.Cause, addr address.Address) {
	snap := c.snapshot()
	prompts := prompt.Build(prompt.Input{
		Cause:        cause,
		SystemPrompt: c.opts.SystemPrompt,
		Window:       c.conv.Window(),
		State:        snap.State,
		Note:         snap.Note,
	})

	c.state.Touch(time.Now())

	tools := GetTools(snap)
	c.busy = true
	go func() {
		out, err := generate(c.ctx, c.llm, prompts, tools)
		c.send(reply{cause: cause, addr: addr, snap: snap, out: out, err: err})
	}()
}

func (c *Core) finishReply(r reply) {
	var typ string
	switch r.cause {
	case prompt.Passive:
		typ = "passive"
	case prompt.Initiative:
		typ = "initiative"
	}

	if r.err != nil {
		ylog.Error("Process <%s> failed: %s", typ, r.err)
		c.deliver(errorAction(r.err, r.addr))
		return
	}
	c.applyOutput(r.out)
	debugOutput(r.out.Answer, r.snap, c.snapshot())

	// After processed, store the assistant message
	snap := c.snapshot()

	assistantMessage := event.AssistantMessage{
		Meta:       c.newMeta(snap.State),
		Content:    r.out.Answer,
		Address:    r.addr,
		Initiative: r.cause == prompt.Initiative,
		Reply:      r.out.Reply,
	}

	if err := c.record(assistantMessage); err != nil {
		ylog.Error("<%s reply> history appended error: %s", typ, err)
		c.deliver(errorAction(yerror.ErrMessageStore, r.addr))
		return
	}

	ylog.Success("Generated %s output %#v", typ, r.out.Answer)

	if c.checkTokenThreshold(70, 0) {
		ylog.Warn("Context too big.(70%%) Start a async compression task.")
		c.queue = append([]job{compressJob{}}, c.queue...)
	}

	c.deliver(&action.AssistantMessage{
		Address:    assistantMessage.Address,
		Content:    assistantMessage.Content,
		Timestamp:  assistantMessage.Timestamp,
		Initiative: assistantMessage.Initiative,
		Reply:      assistantMessage.Reply,
	})
}

func errorAction(err error, addr address.Address) action.Action {
	code := action.ErrMessageProcess

	switch {
	case errors.Is(err, yerror.ErrLLMChoice):
		code = action.ErrLLMChoiceEmpty
	case errors.Is(err, yerror.ErrLLMGenerate):
		code = action.ErrLLMGenerate
	case errors.Is(err, yerror.ErrMessageStore):
		code = action.ErrMessageStore
	}

	return &action.Error{
		Code:    code,
		Message: err.Error(),
		Address: addr,
	}
}

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
	c.state.Apply(out.Change)

	if out.Note != keepNote {
		c.conv.SetNote(out.Note)
	}
}
