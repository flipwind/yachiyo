package gateway

import (
	"context"
	"sync"
	"yachiyo/yachiyo-shared/action"
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

	h.mu.RLock()
	s, ok := h.routes[scheme]
	h.mu.RUnlock()

	if !ok {
		ylog.Error("Scheme <%v> not found, dropped action <%v>", scheme, a)
		return
	}

	s.Send(a)
}
