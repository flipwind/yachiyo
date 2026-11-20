package core

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Options struct {
	SystemPrompt   string
	CompressPrompt string
	Personality    state.BigFive
	Energy         state.Energy
}

type Deps struct {
	Store      store.Store
	LLM        llm.Client
	Deliver    func(a action.Action)
	Initiative InitiativePolicy //exhaustruct:optional
}

type Core struct {
	opts  Options
	llm   llm.Client
	store store.Store
	conv  *conversation.Conversation
	state *state.State

	inbox      chan trigger.Trigger
	stopped    chan struct{}
	deliver    func(a action.Action)
	initiative InitiativePolicy

	results chan result
	queue   []job //exhaustruct:optional
	busy    bool  //exhaustruct:optional

	ctx context.Context //exhaustruct:optional
}

func New(opts Options, deps Deps) (*Core, error) {
	if deps.LLM == nil {
		return nil, errors.New("core: LLM is required")
	}
	if deps.Store == nil {
		return nil, errors.New("core: Store is required")
	}
	if deps.Deliver == nil {
		return nil, errors.New("core: Deliver is required")
	}
	policy := deps.Initiative
	if policy == nil {
		policy = NeverInitiative{}
		ylog.Warn("core initiative policy not set, default to NeverInitiative")
	}

	var core *Core

	restored, err := deps.Store.Load()
	if err != nil {
		return nil, fmt.Errorf("Storage restored error: %w", err)
	}

	snap := state.Initial(opts.Personality, time.Now())
	elapsed := time.Duration(0)
	note := ""
	var sequence int64
	if cp := restored.Checkpoint; cp != nil {
		snap = cp.State
		elapsed = time.Since(cp.SavedAt)
		note = cp.Note
		sequence = cp.Sequence
	}

	conv := conversation.Restore(restored.Events, restored.Summary, note, sequence)
	ylog.Success("Now history sequence at %v.", conv.Sequence())
	st := state.Restore(opts.Personality, opts.Energy, snap)
	if err := st.Advance(elapsed, time.Now()); err != nil {
		return nil, fmt.Errorf("State advanced error: %w", err)
	}

	core = &Core{
		opts:  opts,
		llm:   deps.LLM,
		store: deps.Store,
		conv:  conv,
		state: st,

		results:    make(chan result, 1),
		inbox:      make(chan trigger.Trigger, 2048),
		stopped:    make(chan struct{}),
		deliver:    deps.Deliver,
		initiative: policy,
	}

	ylog.Success("Successfully restored/created checkpoint and histories.")

	return core, nil
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

func (c *Core) handle(t trigger.Trigger) {
	switch m := t.(type) {
	case *trigger.UserMessage:
		c.queue = append(c.queue, userJob{msg: *m, compressed: false})
	case *trigger.RuntimeStateRequest:
		c.deliver(c.processRuntimeStateRequest(*m))
	case *trigger.MessageHistoryRequest:
		c.deliver(c.processMessageHistoryRequest(*m))
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

func (c *Core) saveCheckpoint() {
	if err := c.checkpoint(); err != nil {
		ylog.Error("Checkpoint saving error: %v", err)
	}
}

func (c *Core) Dispatch(t trigger.Trigger) {
	select {
	case c.inbox <- t:
	case <-c.stopped:
	}
}

func (c *Core) internalState() string {
	msg := fmt.Sprintf("Received timetick %s\n", time.Now().Format("15:04:05"))

	snap := c.snapshot().State
	for _, s := range snap.Drives.All() {
		msg += fmt.Sprintf("State: %v at %v, is %v\n", s.Name, s.Drive.Value, s.Drive.String())
	}
	msg += fmt.Sprintf("Energy: %v\n", snap.Energy)

	return msg
}

func (c *Core) record(e event.Event) error {
	if err := c.store.Append(e); err != nil {
		return fmt.Errorf("Storage saving event error: %w", err)
	}

	c.conv.Append(e)
	return nil
}

func (c *Core) commitSummary(summary conversation.Summary) error {
	if err := c.store.SaveSummary(summary); err != nil {
		return fmt.Errorf("Storage saving summary error: %w", err)
	}

	c.conv.SetSummary(summary)
	return nil
}

func (c *Core) newMeta(snap state.Snapshot) event.Meta {
	return event.Meta{
		ID:        uuid.NewV7(),
		Sequence:  c.conv.NextSequence(),
		Timestamp: time.Now(),
		Snapshot:  snap,
	}
}
