package core

import (
	"fmt"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/yerror"
)

type job interface {
	isJob()
}

type userJob struct {
	msg        trigger.UserMessage
	compressed bool
}

func (userJob) isJob() {}

type initiativeJob struct {
	compressed bool
}

func (initiativeJob) isJob() {}

type compressJob struct{}

func (compressJob) isJob() {}

type result interface {
	isResult()
}

type reply struct {
	cause prompt.Cause
	addr  address.Address
	snap  Snapshot
	out   output
	err   error
}

func (reply) isResult() {}

type compressed struct {
	summary conversation.Summary
	err     error
}

func (compressed) isResult() {}

func (c *Core) pump() {
	for !c.busy && len(c.queue) > 0 {
		j := c.queue[0]
		c.queue = c.queue[1:]
		c.start(j)
	}
}

func (c *Core) start(j job) {
	switch task := j.(type) {
	case userJob:
		c.startUser(task)
	case initiativeJob:
		c.startInitiative(task)
	case compressJob:
		c.startCompress()
	}
}

func (c *Core) finish(r result) {
	c.busy = false
	switch task := r.(type) {
	case reply:
		c.finishReply(task)
	case compressed:
		c.finishCompress(task)
	}
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

func (c *Core) startCompress() {
	window := c.conv.Window()

	if len(window.Events) == 0 {
		ylog.Warn("Current history list is empty. Skip compress.")
		return
	}

	var parent uuid.UUID
	if window.Summary == nil {
		parent = uuid.Nil()
	} else {
		parent = window.Summary.ID
	}

	summary := conversation.Summary{
		ID:          uuid.NewV7(),
		CreateAt:    time.Now(),
		Parent:      parent,
		FromEventID: window.Events[0].GetMeta().ID,
		ToEventID:   window.Events[len(window.Events)-1].GetMeta().ID,
		Content:     "NO COMPRESS RESULT",
	}

	messages := prompt.BuildCompression(window, c.opts.CompressPrompt)

	c.busy = true
	go func() {
		sum, err := summarize(c.ctx, c.llm, messages)
		summary.Content = sum
		c.send(compressed{summary: summary, err: err})
	}()
}

func (c *Core) send(r result) {
	select {
	case c.results <- r:
	case <-c.ctx.Done():
	}
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

func (c *Core) finishCompress(r compressed) {
	if r.err != nil {
		ylog.Error("Process compression failed: %s", r.err)
		return
	}

	if err := c.commitSummary(r.summary); err != nil {
		saveErr := fmt.Errorf("Compressed but failed to save in storage: %w", err)
		ylog.Error("Compressed failed: %v", saveErr)
		return
	}

	ylog.Success("Compressed.")
}
