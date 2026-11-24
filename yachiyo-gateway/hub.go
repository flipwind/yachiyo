package gateway

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/endpoint"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Gateway")

type Hub struct {
	outbox chan action.Action // from server to the hub

	routes map[string]Service // distribute the actions
	mu     sync.RWMutex       //exhaustruct:optional
}

func NewHub() *Hub {
	return &Hub{
		outbox: make(chan action.Action, 2048),
		routes: make(map[string]Service),
	}
}

func (h *Hub) lookup(scheme string) (Service, error) {
	h.mu.RLock()
	s, ok := h.routes[scheme]
	h.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("Scheme <%v> not found", scheme)
	}

	return s, nil
}

func (h *Hub) Attach(s Service, port int64, submit func(trig trigger.Trigger)) {
	h.register(s)
	s.Listen(port, submit)
}

func (h *Hub) register(s Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.routes[s.SchemeName()] = s
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case a := <-h.outbox:
			h.route(a)
		}
	}
}

func (h *Hub) Deliver(a action.Action) {
	select {
	case h.outbox <- a:
	default:
		ylog.Warn("Outbox fulled. Drop action <%v>.", a)
	}
}

func (h *Hub) route(a action.Action) {
	scheme, err := a.GetAddress().Scheme()
	if err != nil {
		ylog.Error("Action <%v> route error: %v", a, err)
		return
	}

	s, err := h.lookup(scheme)
	if err != nil {
		ylog.Error("Drop action <%T>: %v", a, err)
		return
	}

	if err := s.Send(a); err != nil {
		ylog.Error("Send action <%v> failed: %v", a, err)
	}
}

func (h *Hub) AvailableAddresses(cap string) ([]address.Address, error) {
	if !slices.Contains(endpoint.AvailableCapability(), cap) {
		return []address.Address{}, errors.New("capability not defined")
	}

	addrs := make([]address.Address, 0)

	h.mu.RLock()
	routes := maps.Clone(h.routes)
	h.mu.RUnlock()

	for _, service := range routes {
		if provider, ok := service.(EndpointProvider); ok {
			for _, client := range provider.Clients() {
				if client.Capability.Has(cap) {
					addrs = append(addrs, client.Address)
				}
			}
		}
	}

	return addrs, nil
}

func (h *Hub) ClientTools() []endpoint.RemoteTool {
	tools := make([]endpoint.RemoteTool, 0)

	h.mu.RLock()
	routes := maps.Clone(h.routes)
	h.mu.RUnlock()

	for _, service := range routes {
		if provider, ok := service.(EndpointProvider); ok {
			for _, client := range provider.Clients() {
				tools = append(tools, slices.Clone(client.Tools)...)
			}
		}
	}

	return tools
}

func (h *Hub) SendMessage(a action.AssistantMessage) error {
	scheme, err := a.Address.Scheme()
	if err != nil {
		return err
	}

	s, err := h.lookup(scheme)
	if err != nil {
		ylog.Error("Drop action <%T>: %v", a, err)
		return err
	}

	return s.Send(&a)
}

func (h *Hub) UseTool(ctx context.Context, addr address.Address, reqID, toolName, payload string) (string, error) {
	scheme, err := addr.Scheme()
	if err != nil {
		return "", err
	}

	s, err := h.lookup(scheme)
	if err != nil {
		return "", fmt.Errorf("Tool failed: %v", err)
	}

	if executor, ok := s.(ToolExecutor); ok {
		return executor.UseTool(ctx, addr, reqID, toolName, payload)
	}
	return "", fmt.Errorf("Service doesn't have an executor method.")
}
