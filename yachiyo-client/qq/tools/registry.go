package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
)

const defaultToolTimeout = 4500 * time.Millisecond

type HandlerFunc func(ctx context.Context, payload json.RawMessage) (string, error)

type Registry struct {
	clientID string
	api      *qq.APIClient
	store    *store.Store
	specs    []yws.ToolItem
	handlers map[string]HandlerFunc
}

func NewRegistry(clientID string, api *qq.APIClient, st *store.Store) *Registry {
	r := &Registry{
		clientID: clientID,
		api:      api,
		store:    st,
		handlers: make(map[string]HandlerFunc),
	}

	r.registerGroupTools()
	r.registerMessageTools()
	return r
}

func (r *Registry) register(spec yws.ToolItem, fn HandlerFunc) {
	r.specs = append(r.specs, spec)
	r.handlers[spec.Name] = fn
}

func (r *Registry) Specs() []yws.ToolItem {
	out := make([]yws.ToolItem, len(r.specs))
	copy(out, r.specs)
	return out
}

func (r *Registry) Execute(ctx context.Context, call yws.ToolCall) (bool, string) {
	fn, ok := r.handlers[call.ToolName]
	if !ok {
		return false, fmt.Sprintf("unknown tool: %s", call.ToolName)
	}

	execCtx, cancel := context.WithTimeout(ctx, defaultToolTimeout)
	defer cancel()

	res, err := fn(execCtx, call.Payload)
	if err != nil {
		return false, err.Error()
	}
	return true, res
}

func marshalJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
