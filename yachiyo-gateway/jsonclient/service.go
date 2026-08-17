package jsonclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/jsonclient/model"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var ylog = logger.New("Yachiyo.Jsonclient")

type JsonClientService struct {
	channel *gateway.GatewayChannel

	clients map[string]*Client
	mutex   sync.RWMutex
}

func NewJsonClientService() *JsonClientService {
	return &JsonClientService{
		clients: make(map[string]*Client),
	}
}

func (s *JsonClientService) Listen(c *gateway.GatewayChannel, p int64) {
	s.channel = c
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", s.handleWebsocket)

	go func() {
		port := fmt.Sprintf(":%d", p)
		ylog.Success("Running client server on :%d successfully.", p)
		if err := http.ListenAndServe(port, mux); err != nil {
			ylog.Error("json client adapter running error: %v", err)
		}
	}()

	go s.ListenSend()
}

func (s *JsonClientService) SchemeName() string {
	return "jsonclient"
}

func (s *JsonClientService) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		ylog.Error("Websocket upgrade error: %v", err)
		return
	}

	client := NewClient(conn)
	go client.Run(s.unregister, s.handleReceive)
}

func (s *JsonClientService) handleReceive(c *Client, data []byte) {
	var message model.Envelope

	if err := json.Unmarshal(data, &message); err != nil {
		ylog.Error("JSON unmarshal error: %v", err)
		return
	}

	switch message.Category {
	case "connection":
		s.handleConnection(c, message)
	case "interaction":
		s.handleInteraction(c, message)
	}
}

func (s *JsonClientService) handleConnection(c *Client, message model.Envelope) {
	switch message.Type {
	case "register":
		var data model.Register
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if data.ClientType != "IM" && data.ClientType != "Client" {
			c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_info_error"})
			return
		}

		var old *Client

		s.mutex.Lock()
		if oldClient, ok := s.clients[data.ClientID]; ok == true {
			if oldClient.Type != data.ClientType {
				c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_conflict"})
				s.mutex.Unlock()
				return
			}
			old = oldClient
		}

		c.Type = data.ClientType
		c.Name = data.ClientName
		c.ID = data.ClientID

		s.clients[data.ClientID] = c
		s.mutex.Unlock()

		if old != nil {
			old.conn.Close(websocket.StatusNormalClosure, "")
		}
		ylog.Success("Registered [%s @%s](%s).", c.Type, c.Name, c.ID)
		c.send("connection", "register_success", &model.RegisterSuccess{RuntimeName: "Yachiyo", RuntimeVersion: "0.1"})
		// TODO: read from config
	case "heartbeat":
		var data model.HeartBeat
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		c.LastHeartbeatTime = time.Now()
	case "offline":
		var data model.Offline
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		s.unregister(c)
	}
}

func (s *JsonClientService) handleInteraction(c *Client, message model.Envelope) {
	switch message.Type {
	case "client_message":
		var data model.ClientMessage
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		s.channel.ToServer <- &trigger.Message{
			Type:     "user",
			Author:   "user",
			Platform: c.Type,
			Content:  data.Message,
			Time:     time.Now().Unix(),

			Address: address.Address{
				Content: fmt.Sprintf("%s://%s", s.SchemeName(), c.ID),
			},
		}
	}
}

func (s *JsonClientService) checkClient(c *Client) bool {
	if c.ID == "" {
		// TODO: unknown client
		return false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	client, ok := s.clients[c.ID];
	
	return ok && client == c
}

func (s *JsonClientService) unregister(c *Client) {
	s.mutex.Lock()
	if client, ok := s.clients[c.ID]; ok && client == c {
		delete(s.clients, c.ID)
	}
	s.mutex.Unlock()

	ylog.Success("Unregistered [%s @%s](%s).", c.Type, c.Name, c.ID)
}

func (s *JsonClientService) ListenSend() {
	for act := range s.channel.ToClient {
		switch t := act.(type) {
		case *action.Message:
			addr := t.Address.Host()
			c := s.clients[addr]
			c.send("interaction", "runtime_message", &model.RuntimeMessage{Message: t.Content, IsInitiative: false})
		default:
			ylog.Info("Unsupport value: %T", t)
		}
	}
}
