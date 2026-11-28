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
	sent  []sentMessage
	err   error
}

func (reply) isResult() {}

type sentMessage struct {
	Address   address.Address
	Content   []string
	Timestamp time.Time
}

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
		Meta:     c.newMeta(snap.State, j.msg.Timestamp),
		Content:  j.msg.Content,
		Author:   j.msg.Author,
		Platform: j.msg.Platform,
		Address:  j.msg.Address,
	}

	if err := c.record(userMessage); err != nil {
		ylog.Error("UserMessage history appended error: %s", err)
		c.hub.Deliver(errorAction(yerror.ErrMessageStore, j.msg.Address))
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
		Nickname:     c.opts.Nickname,
		SystemPrompt: c.opts.SystemPrompt,
		Window:       c.conv.Window(),
		State:        snap.State,
		Note:         snap.Note,
	})

	c.state.Touch(time.Now())

	var sent []sentMessage
	tools := getTools(c.ctx, snap, cause, &sent, c.hub)
	c.busy = true
	go func() {
		out, err := generate(c.ctx, c.llm, prompts, tools)
		c.send(reply{cause: cause, addr: addr, snap: snap, out: out, sent: sent, err: err})
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

	if r.err == nil {
		c.applyOutput(r.out)
	}

	// After processed, store the assistant message
	snap := c.snapshot()
	for _, m := range r.sent {
		if err := c.record(event.AssistantMessage{
			Meta:       c.newMeta(snap.State, m.Timestamp),
			Content:    m.Content,
			Address:    m.Address,
			Initiative: r.cause == prompt.Initiative,
		}); err != nil {
			ylog.Error("<%s reply> history appended error: %s", typ, err)
			c.hub.Deliver(errorAction(yerror.ErrMessageStore, r.addr))
		}
	}

	if r.err != nil {
		ylog.Error("Process <%s> failed: %s", typ, r.err)
		c.hub.Deliver(errorAction(r.err, r.addr))
		return
	}

	debugOutput(r.out.Thoughts, r.snap, c.snapshot())

	ylog.Success("Yachiyo(%s): %#v", typ, r.out.Thoughts)

	if c.checkTokenThreshold(70, 0) {
		ylog.Warn("Context too big.(70%%) Start a async compression task.")
		c.queue = append([]job{compressJob{}}, c.queue...)
	}
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

func debugOutput(thoughts string, c_formal Snapshot, c_later Snapshot) {
	ylog.Debug(`== DEBUG MESSAGE ==
Yachiyo > %v
* Formal:
%s
* Later:
%s
%+v
* Session Context: %s`,
		thoughts, prompt.Drives(c_formal.State.Drives),
		prompt.Drives(c_later.State.Drives), c_later.State.Determination,
		c_later.Note)
}

func (c *Core) applyOutput(out output) {
	c.state.Apply(out.Change)

	if out.Note != keepNote {
		c.conv.SetNote(out.Note)
	}
}
