package jsonclient

import (
	"encoding/json"
	"regexp"
	"time"
	"uuid"
	"yachiyo/yachiyo-gateway/jsonclient/model"
	"yachiyo/yachiyo-gateway/pending"
	"yachiyo/yachiyo-shared/endpoint"
	"yachiyo/yachiyo-shared/trigger"

	"github.com/coder/websocket"
)

func (s *JsonClientService) handleReceive(c *Client, data []byte) {
	var message model.Envelope

	if err := json.Unmarshal(data, &message); err != nil {
		ylog.Error("JSON unmarshal error: %v", err)
		return
	}

	registered := (c.ID != "")

	// unregister client
	if !registered && !(message.Category == "connection" && message.Type == "register") {
		c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_unknown"})
		return
	}

	switch message.Category {
	case "connection":
		s.handleConnection(c, message)
	case "interaction":
		s.handleInteraction(c, message)
	case "state":
		s.handleState(c, message)
	case "tool":
		s.handleTools(c, message)
	}
}

var validClientName = regexp.MustCompile(`^[a-zA-Z0-9_]{1,16}$`)

func (s *JsonClientService) handleConnection(c *Client, message model.Envelope) {
	switch message.Type {
	case "register":
		if c.ID != "" {
			c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_already_registered"})
			return
		}

		var data model.Register
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if !validClientName.MatchString(data.ClientName) {
			c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_name_invalid"})
			return
		}

		if data.ClientType != "IM" && data.ClientType != "Client" {
			c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_info_error"})
			return
		}

		clientUUID, err := uuid.Parse(data.ClientID)
		if err != nil {
			c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_id_invalid"})
			return
		}

		var clientID = clientUUID.String()

		var old *Client

		s.mu.Lock()
		if oldClient, ok := s.clients[clientID]; ok == true {
			// oldClient exists; register.
			if oldClient.Type != data.ClientType {
				c.send("connection", "register_error", &model.RegisterError{ErrorType: "client_conflict"})
				s.mu.Unlock()
				return
			}
			old = oldClient
		}

		if old != nil && old != c {
			// The same clientID has two client now
			old.conn.Close(websocket.StatusNormalClosure, "")
			delete(s.clients, old.ID)
		}

		c.Type = data.ClientType
		c.Name = data.ClientName
		c.ID = clientID
		c.LastHeartbeatTime = time.Now()

		s.clients[clientID] = c
		s.mu.Unlock()

		ylog.Success("Registered [%s @%s](%s).", c.Type, c.Name, c.ID)

		c.send("connection", "register_success",
			&model.RegisterSuccess{RuntimeName: s.serverInfo.Name, RuntimeVersion: s.serverInfo.Version})
	case "heartbeat":
		var data model.HeartBeat
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		s.mu.Lock()
		c.LastHeartbeatTime = time.Now()
		s.mu.Unlock()

		c.send("connection", "heartbeat_respond", &model.HeartBeatRespond{})
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

		s.submit(&trigger.UserMessage{
			Author:    "user",
			Platform:  c.Type,
			Content:   data.Message,
			Timestamp: time.Now(),

			Address: s.clientAddress(c),
		})
	case "get_relative_message_history":
		if s.checkClient(c) == false {
			return
		}
		s.submit(&trigger.MessageHistoryRequest{
			Address: s.clientAddress(c),
		})
	}
}

func (s *JsonClientService) handleState(c *Client, message model.Envelope) {
	switch message.Type {
	case "runtime_state_request":
		s.submit(&trigger.RuntimeStateRequest{
			Address: s.clientAddress(c),
		})
	}
}

func (s *JsonClientService) handleTools(c *Client, message model.Envelope) {
	switch message.Type {
	case "update_tools":
		newTools := make([]endpoint.RemoteTool, 0)
		newCapability := endpoint.NewCapability()

		var data model.UpdateTools
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		reject_capability := make([]string, 0)
		reject_tool := make(map[string]string)

		seen := make(map[string]bool)
		canbe := make(map[string]bool)
		for _, cap := range data.Capability {
			if !newCapability.Set(cap) {
				reject_capability = append(reject_capability, cap)
				continue
			}
		}
		for _, t := range data.Tools {
			ok := t.ValidName()
			if !ok {
				reject_tool[t.Name] = "name_error"
				continue
			}
			if seen[t.Name] {
				canbe[t.Name] = false
				reject_tool[t.Name] = "duplicate_name"
				continue
			}
			canbe[t.Name] = true
			seen[t.Name] = true
		}
		for _, t := range data.Tools {
			if !canbe[t.Name] {
				continue
			}

			newTools = append(newTools, endpoint.RemoteTool{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
				Address:     s.clientAddress(c),
			})
		}

		s.mu.Lock()
		c.Capability = newCapability
		c.Tools = newTools
		s.mu.Unlock()

		if len(reject_capability) != 0 || len(reject_tool) != 0 {
			c.send("tool", "update_tools_feedback", &model.UpdateToolsFeedback{
				Status:           "error",
				RejectTools:      reject_tool,
				RejectCapability: reject_capability,
			})
			return
		}

		ylog.Success("Client <%s> now have %v capabilities and %v tools.", c.Name, newCapability.EnableCounts(), len(newTools))

		c.send("tool", "update_tools_feedback", &model.UpdateToolsFeedback{
			Status:           "success",
			RejectTools:      make(map[string]string),
			RejectCapability: make([]string, 0),
		})
	case "tool_call_result":
		var data model.ToolCallResult
		if err := json.Unmarshal(message.Data, &data); err != nil {
			ylog.Error("JSON unmarshal error: %v", err)
			return
		}

		if s.checkClient(c) == false {
			return
		}

		s.pending.Resolve(data.RequestID, pending.Result{
			OK:     data.OK,
			Result: data.Result,
		})
	}
}
