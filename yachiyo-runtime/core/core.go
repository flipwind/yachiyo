package core

import (
	"context"
	"errors"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

// Options defines a series of data, usually read only once.
type Options struct {
	SystemPrompt   string
	CompressPrompt string
	Personality    state.BigFive
	Energy         state.Energy
}

// Deps, which is also called dependencies, holds some usable
// behavior components.
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

func (c *Core) Dispatch(t trigger.Trigger) {
	select {
	case c.inbox <- t:
	case <-c.stopped:
	}
}
