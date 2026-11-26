package wsjson

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"

	"github.com/coder/websocket"
)

func (s *Service) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		ylog.Error("Websocket upgrade error: %v", err)
		return
	}

	client := NewClient(conn)
	go client.Run(s.unregister, s.handleReceive)
}

const maxPathLength = 128

var validPathChars = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

func normalizePath(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	trimmed := strings.Trim(raw, "/")
	if trimmed == "" {
		return "", errors.New("path must not only consist `/`s")
	}
	if len(trimmed) > maxPathLength {
		return "", fmt.Errorf("path exceeds maximum length of %d", maxPathLength)
	}
	if !validPathChars.MatchString(trimmed) {
		return "", errors.New("path contains invalid characters (allowed: a-z, A-Z, 0-9, ., _, -, /)")
	}

	for seg := range strings.SplitSeq(trimmed, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", fmt.Errorf("path contains invalid segment %q", seg)
		}
	}

	return trimmed, nil
}

func (s *Service) clientAddress(c *Client) address.Address {
	return address.New(fmt.Sprintf("%s://%s", s.SchemeName(), c.ID))
}

func (s *Service) clientAddressWithPath(c *Client, path string) address.Address {
	if path == "" {
		return s.clientAddress(c)
	}
	return address.New(fmt.Sprintf("%s://%s/%s", s.SchemeName(), c.ID, path))
}

func (s *Service) getClient(a action.Action) (*Client, error) {
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
func (s *Service) checkClient(c *Client) bool {
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

func (s *Service) heartbeatCleanup() {
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

func (s *Service) unregister(c *Client) {
	s.mu.Lock()
	if client, ok := s.clients[c.ID]; ok && client == c {
		delete(s.clients, c.ID)
	}
	s.mu.Unlock()

	c.conn.Close(websocket.StatusAbnormalClosure, "")

	ylog.Success("Unregistered [%s @%s](%s).", c.Type, c.Name, c.ID)
}
