package wsjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"
	"yachiyo/yachiyo-gateway/wsjson/model"
	"yachiyo/yachiyo-gateway/pending"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/endpoint"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.wsjson")

type Service struct {
	clients map[string]*Client
	mu      sync.RWMutex //exhaustruct:optional

	submit     func(trig trigger.Trigger) //exhaustruct:optional
	pending    *pending.Pending
	serverInfo ServerInfo
}

func New(info ServerInfo) *Service {
	return &Service{
		clients:    make(map[string]*Client),
		serverInfo: info,
		pending:    pending.NewPending(),
	}
}

func (s *Service) Listen(port int64, submit func(trig trigger.Trigger)) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", s.handleWebsocket)

	s.submit = submit

	go func() {
		addr := fmt.Sprintf(":%d", port)
		ylog.Success("Running wsjson server on :%d successfully.", port)
		if err := http.ListenAndServe(addr, mux); err != nil {
			ylog.Error("wsjson gateway running error: %v", err)
		}
	}()
	go s.heartbeatCleanup()
}

func (s *Service) SchemeName() string {
	return "wsjson"
}

func (s *Service) Send(a action.Action) error {
	switch t := a.(type) {
	case *action.AssistantMessage:
		c, err := s.getClient(t)
		if err != nil {
			ylog.Error("Process action failed: %v", err)
			return err
		}

		if t.Initiative == true && t.Reply == false {
			return nil
		}

		s.mu.RLock()
		can := c.Capability.Has(endpoint.CapMessageText)
		s.mu.RUnlock()
		if !can {
			return fmt.Errorf("Client <%s> doesn't support %s", c.ID, endpoint.CapMessageText)
		}

		return c.send("interaction", "runtime_message", &model.RuntimeMessage{Time: t.Timestamp.Unix(), Reply: t.Reply, Message: t.Content, IsInitiative: t.Initiative})
	case *action.RuntimeState:
		c, err := s.getClient(t)
		if err != nil {
			ylog.Error("Process action failed: %v", err)
			return err
		}

		actives := make([]model.OCCEmotion, 0)
		for _, emo := range t.Affect.Actives {
			actives = append(actives, model.OCCEmotion{
				Type:      emo.Type,
				Intensity: emo.Intensity,
			})
		}

		return c.send("state", "runtime_state", &model.RuntimeState{
			Timetick: t.Timetick,
			Energy:   t.Energy,
			Affect: struct {
				Octants   string             `json:"octants"`
				Pleasure  float64            `json:"pleasure"`
				Arousal   float64            `json:"arousal"`
				Dominance float64            `json:"dominance"`
				Actives   []model.OCCEmotion `json:"actives"`
			}{
				Octants:   t.Affect.Octants,
				Pleasure:  t.Affect.Pleasure,
				Arousal:   t.Affect.Arousal,
				Dominance: t.Affect.Dominance,
				Actives:   actives,
			},
		})
	case *action.MessageHistory:
		c, err := s.getClient(t)
		if err != nil {
			ylog.Error("Process action failed: %v", err)
			return err
		}

		msgs := make([]model.Envelope, 0)
		for _, msg := range t.Messages {
			switch m := msg.(type) {
			case *action.AssistantHistoryMessage:
				data, err := json.Marshal(model.RuntimeMessage{
					Reply:        m.Reply,
					IsInitiative: m.Initiative,
					Message:      m.Content,
					Time:         m.Timestamp.Unix(),
				})
				if err != nil {
					ylog.Error("AssistantMessage Marshal error: %v", err)
				}

				msgs = append(msgs, model.Envelope{
					Category: "interaction",
					Type:     "runtime_message",
					Data:     data,
				})
			case *action.UserHistoryMessage:
				data, err := json.Marshal(model.ClientMessage{
					Message: m.Content,
					Time:    m.Timestamp.Unix(),
				})
				if err != nil {
					ylog.Error("AssistantMessage Marshal error: %v", err)
				}

				msgs = append(msgs, model.Envelope{
					Category: "interaction",
					Type:     "client_message",
					Data:     data,
				})
			}
		}

		return c.send("interaction", "relative_message_history", &model.RelativeMessageHistory{Messages: msgs})
	case *action.Error:
		c, err := s.getClient(t)
		if err != nil {
			ylog.Error("Process action failed: %v", err)
			return err
		}

		return c.send("interaction", "runtime_error", &model.RuntimeError{
			Code:    string(t.Code),
			Message: t.Message,
		})
	default:
		ylog.Info("Unsupport value: %T", t)
		return fmt.Errorf("Unsupported action type: %T", t)
	}
}

func (s *Service) UseTool(ctx context.Context, addr address.Address, reqID, toolName, payload string) (string, error) {
	host, err := addr.Host()
	if err != nil {
		return "", fmt.Errorf("%s: %w", s.SchemeName(), err)
	}

	s.mu.RLock()
	c, ok := s.clients[host]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("%s: client %s not found", s.SchemeName(), host)
	}

	s.pending.Register(reqID)

	pload := json.RawMessage("{}")
	if json.Valid([]byte(payload)) {
		pload = json.RawMessage(payload)
	}

	if err := c.send("tool", "tool_call", &model.ToolCall{
		ToolName:  toolName,
		RequestID: reqID,
		Payload:   pload,
	}); err != nil {
		s.pending.Remove(reqID)
		return "", err
	}

	res, err := s.pending.Wait(ctx, reqID, time.Second*5)
	if err != nil {
		return "", fmt.Errorf("failed to get result: %w", err)
	}
	if !res.OK {
		return "", errors.New(res.Result)
	}

	return res.Result, nil
}

func (s *Service) Clients() []endpoint.Endpoint {
	li := make([]endpoint.Endpoint, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.clients {
		li = append(li, endpoint.Endpoint{
			Name:       c.Name,
			ID:         c.ID,
			Address:    s.clientAddress(c),
			Capability: c.Capability.Clone(),
			Tools:      slices.Clone(c.Tools),
		})
	}
	return li
}
