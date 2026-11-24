package jsonclient

import (
	"fmt"
	"net/http"
	"time"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"

	"github.com/coder/websocket"
)

func (s *JsonClientService) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		ylog.Error("Websocket upgrade error: %v", err)
		return
	}

	client := NewClient(conn)
	go client.Run(s.unregister, s.handleReceive)
}

func (s *JsonClientService) clientAddress(c *Client) address.Address {
	return address.New(fmt.Sprintf("%s://%s", s.SchemeName(), c.ID))
}

func (s *JsonClientService) getClient(a action.Action) (*Client, error) {
	address := a.GetAddress()
	host, err := address.Host()
	if err != nil {
		return nil, fmt.Errorf("Action address error: %w", err)
	}

	s.mu.RLock()
	c := s.clients[host]
	s.mu.RUnlock()

	if c == nil {
		return nil, fmt.Errorf("Unknown address: %s", host)
	}

	return c, nil
}

// Check client if former client is not the same as the current one.
func (s *JsonClientService) checkClient(c *Client) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, ok := s.clients[c.ID]

	return ok && client == c
}

// ServerInfo is what the client learns from the runtime when it registers.
type ServerInfo struct {
	Name    string
	Version string
}

func (s *JsonClientService) heartbeatCleanup() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var expired []*Client
		nowtime := time.Now()

		s.mu.RLock()
		for _, client := range s.clients {
			if nowtime.Sub(client.LastHeartbeatTime) > 30*time.Second {
				expired = append(expired, client)
			}
		}
		s.mu.RUnlock()

		for _, client := range expired {
			s.unregister(client)
		}
	}
}

func (s *JsonClientService) unregister(c *Client) {
	s.mu.Lock()
	if client, ok := s.clients[c.ID]; ok && client == c {
		delete(s.clients, c.ID)
	}
	s.mu.Unlock()

	c.conn.Close(websocket.StatusAbnormalClosure, "")

	ylog.Success("Unregistered [%s @%s](%s).", c.Type, c.Name, c.ID)
}
