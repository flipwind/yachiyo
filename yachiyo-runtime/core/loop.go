package core

import (
	"context"
	"time"
	"yachiyo/yachiyo-shared/trigger"
)

type job interface {
	isJob()
}

type result interface {
	isResult()
}

func (c *Core) Run(ctx context.Context) {
	c.ctx = ctx
	defer close(c.stopped)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.saveCheckpoint()
			return
		case t := <-c.inbox:
			c.handle(t)
		case <-ticker.C:
			c.tick(time.Second)
		case t := <-c.results:
			c.finish(t)
		}
		c.pump()
	}
}

// pump checks if the core's queue is free, and *pump* a job
// to start immediately when the core is not LLM-busy(long tasks).
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

func (c *Core) send(r result) {
	select {
	case c.results <- r:
	case <-c.ctx.Done():
	}
}

func (c *Core) handle(t trigger.Trigger) {
	switch m := t.(type) {
	case *trigger.UserMessage:
		c.queue = append(c.queue, userJob{msg: *m, compressed: false})
	case *trigger.RuntimeStateRequest:
		c.hub.Deliver(c.processRuntimeStateRequest(*m))
	case *trigger.MessageHistoryRequest:
		c.hub.Deliver(c.processMessageHistoryRequest(*m))
	default:
		ylog.Error("Process unsupported type: %T", m)
	}
}

func (c *Core) tick(d time.Duration) {
	if err := c.state.Advance(d, time.Now()); err != nil {
		ylog.Error("State advanced error: %v", err)
	}

	c.saveCheckpoint()

	ylog.Debug("%s", c.internalState())

	snap := c.state.Snapshot()
	now := time.Now()

	if !c.busy && c.initiative.ShouldInitiative(snap, now) {
		ylog.Info("Initiative triggered.")
		c.state.Relieve(d)
		c.queue = append(c.queue, initiativeJob{})
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
