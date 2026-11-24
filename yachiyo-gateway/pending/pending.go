package pending

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Result struct {
	OK      bool
	Result string
}

type Pending struct {
	mu      sync.Mutex //exhaustruct:optional
	waiters map[string]chan Result
}

func NewPending() *Pending {
	return &Pending{
		waiters: make(map[string]chan Result),
	}
}

// Register() is the must step to call.
// After registered, you *must* call Remove(reqID) or Wait(reqID).
func (p *Pending) Register(reqID string) {
	p.mu.Lock()
	p.waiters[reqID] = make(chan Result, 1)
	p.mu.Unlock()
}

func (p *Pending) Remove(reqID string) {
	p.mu.Lock()
	delete(p.waiters, reqID)
	p.mu.Unlock()
}

func (p *Pending) Resolve(reqID string, res Result) {
	p.mu.Lock()
	ch, ok := p.waiters[reqID]
	p.mu.Unlock()

	if ok {
		select {
		case ch <- res:
		default:
		}
	}
}

func (p *Pending) Wait(ctx context.Context, reqID string, timeout time.Duration) (Result, error) {
	defer p.Remove(reqID)

	p.mu.Lock()
	ch, ok := p.waiters[reqID]
	p.mu.Unlock()

	if !ok {
		return Result{}, errors.New("reqID not registered")
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case r := <-ch:
		return r, nil
	case <-timer.C:
		return Result{}, errors.New("timeout")
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}
