package qq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var wsLog = logger.New("Yachiyo.QQ.Gateway")

type EventHandlers struct {
	OnReady           func(ready ReadyData)
	OnGroupMessage    func(ctx context.Context, ev GroupMessageEvent, isAtEvent bool)
	OnGroupAddRobot   func(ev GroupRobotEvent)
	OnGroupDelRobot   func(ev GroupRobotEvent)
	OnGroupMsgReject  func(ev GroupRobotEvent)
	OnGroupMsgReceive func(ev GroupRobotEvent)
	OnC2CMessage      func(ctx context.Context, ev C2CMessageEvent)
	OnFriendAdd       func(ev FriendEvent)
	OnFriendDel       func(ev FriendEvent)
	OnC2CMsgReject    func(ev FriendEvent)
	OnC2CMsgReceive   func(ev FriendEvent)
}

type GatewayClient struct {
	api      *APIClient
	tokens   *TokenProvider
	handlers EventHandlers

	mu        sync.Mutex
	sessionID string
	lastSeq   atomic.Int64
	hasSeq    atomic.Bool
}

func NewGatewayClient(api *APIClient, tokens *TokenProvider, handlers EventHandlers) *GatewayClient {
	return &GatewayClient{api: api, tokens: tokens, handlers: handlers}
}

func (g *GatewayClient) Run(ctx context.Context) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		err := g.runOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		wsLog.Warn("QQ Gateway disconnected: %v. Reconnecting in %s...", err, backoff)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (g *GatewayClient) runOnce(ctx context.Context) error {
	gwURL, err := g.api.GetGatewayURL(ctx)
	if err != nil {
		return fmt.Errorf("fetch gateway url: %w", err)
	}
	token, err := g.tokens.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("fetch access_token: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	conn, _, err := websocket.Dial(dialCtx, gwURL, nil)
	cancel()
	if err != nil {
		return fmt.Errorf("dial qq gateway: %w", err)
	}
	conn.SetReadLimit(16 << 20)
	defer conn.Close(websocket.StatusNormalClosure, "")

	var writeMu sync.Mutex
	sendJSON := func(sendCtx context.Context, v any) error {
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.Write(sendCtx, websocket.MessageText, raw)
	}

	sessionCtx, sessionCancel := context.WithCancel(ctx)
	defer sessionCancel()

	var hbStarted bool
	for {
		_, raw, err := conn.Read(sessionCtx)
		if err != nil {
			return err
		}
		var payload GatewayPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			wsLog.Error("Decode QQ gateway payload error: %v", err)
			continue
		}
		if payload.S != nil {
			g.lastSeq.Store(*payload.S)
			g.hasSeq.Store(true)
		}

		switch payload.Op {
		case OpHello:
			var hello HelloData
			if err := json.Unmarshal(payload.D, &hello); err != nil {
				return fmt.Errorf("decode hello: %w", err)
			}
			interval := time.Duration(hello.HeartbeatInterval) * time.Millisecond
			if interval <= 0 {
				interval = 30 * time.Second
			}
			if !hbStarted {
				hbStarted = true
				go g.heartbeatLoop(sessionCtx, interval, sendJSON)
			}
			g.mu.Lock()
			sid := g.sessionID
			g.mu.Unlock()
			if sid != "" && g.hasSeq.Load() {
				wsLog.Info("Resuming QQ Gateway session %s (seq=%d)...", sid, g.lastSeq.Load())
				err = sendJSON(sessionCtx, map[string]any{"op": OpResume, "d": ResumeData{Token: "QQBot " + token, SessionID: sid, Seq: g.lastSeq.Load()}})
			} else {
				wsLog.Info("Identifying with QQ Gateway (intents=%d)...", IntentGroupAndC2CEvent)
				err = sendJSON(sessionCtx, map[string]any{"op": OpIdentify, "d": IdentifyData{Token: "QQBot " + token, Intents: IntentGroupAndC2CEvent, Shard: [2]int{0, 1}}})
			}
			if err != nil {
				return fmt.Errorf("send auth frame: %w", err)
			}
		case OpHeartbeatACK:
			wsLog.Debug("QQ Gateway heartbeat ACK")
		case OpReconnect:
			wsLog.Warn("QQ Gateway requested reconnect (op=7)")
			return fmt.Errorf("server requested reconnect")
		case OpInvalidSession:
			wsLog.Warn("QQ Gateway invalid session (op=9), resetting session_id")
			g.mu.Lock()
			g.sessionID = ""
			g.mu.Unlock()
			g.hasSeq.Store(false)
			return fmt.Errorf("invalid session")
		case OpDispatch:
			g.handleDispatch(sessionCtx, payload.T, payload.D)
		}
	}
}

func (g *GatewayClient) heartbeatLoop(ctx context.Context, interval time.Duration, sendJSON func(context.Context, any) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var seq any
			if g.hasSeq.Load() {
				seq = g.lastSeq.Load()
			}
			hbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := sendJSON(hbCtx, map[string]any{"op": OpHeartbeat, "d": seq})
			cancel()
			if err != nil {
				wsLog.Warn("QQ Gateway heartbeat send error: %v", err)
				return
			}
		}
	}
}

func dispatchEvent[T any](data json.RawMessage, name string, fn func(T)) {
	if fn == nil {
		return
	}
	var ev T
	if err := json.Unmarshal(data, &ev); err != nil {
		wsLog.Error("Decode %s error: %v", name, err)
		return
	}
	fn(ev)
}

func (g *GatewayClient) handleDispatch(ctx context.Context, eventType string, data json.RawMessage) {
	switch eventType {
	case "READY":
		dispatchEvent(data, eventType, func(ready ReadyData) {
			g.mu.Lock()
			g.sessionID = ready.SessionID
			g.mu.Unlock()
			wsLog.Success("QQ Gateway READY: bot=%s (%s), session=%s", ready.User.Username, ready.User.ID, ready.SessionID)
			if g.handlers.OnReady != nil {
				g.handlers.OnReady(ready)
			}
		})
	case "RESUMED":
		wsLog.Success("QQ Gateway session resumed")
	case "GROUP_MESSAGE_CREATE":
		dispatchEvent(data, eventType, func(ev GroupMessageEvent) {
			if g.handlers.OnGroupMessage != nil {
				g.handlers.OnGroupMessage(ctx, ev, false)
			}
		})
	case "GROUP_AT_MESSAGE_CREATE":
		dispatchEvent(data, eventType, func(ev GroupMessageEvent) {
			if g.handlers.OnGroupMessage != nil {
				g.handlers.OnGroupMessage(ctx, ev, true)
			}
		})
	case "GROUP_ADD_ROBOT":
		dispatchEvent(data, eventType, func(ev GroupRobotEvent) {
			wsLog.Info("Bot added to group %s", ev.GroupOpenID)
			if g.handlers.OnGroupAddRobot != nil {
				g.handlers.OnGroupAddRobot(ev)
			}
		})
	case "GROUP_DEL_ROBOT":
		dispatchEvent(data, eventType, func(ev GroupRobotEvent) {
			wsLog.Info("Bot removed from group %s", ev.GroupOpenID)
			if g.handlers.OnGroupDelRobot != nil {
				g.handlers.OnGroupDelRobot(ev)
			}
		})
	case "GROUP_MSG_REJECT":
		dispatchEvent(data, eventType, g.handlers.OnGroupMsgReject)
	case "GROUP_MSG_RECEIVE":
		dispatchEvent(data, eventType, g.handlers.OnGroupMsgReceive)
	case "C2C_MESSAGE_CREATE":
		dispatchEvent(data, eventType, func(ev C2CMessageEvent) {
			if g.handlers.OnC2CMessage != nil {
				g.handlers.OnC2CMessage(ctx, ev)
			}
		})
	case "FRIEND_ADD":
		dispatchEvent(data, eventType, func(ev FriendEvent) {
			wsLog.Info("Friend added: %s", ev.OpenID)
			if g.handlers.OnFriendAdd != nil {
				g.handlers.OnFriendAdd(ev)
			}
		})
	case "FRIEND_DEL":
		dispatchEvent(data, eventType, func(ev FriendEvent) {
			wsLog.Info("Friend removed: %s", ev.OpenID)
			if g.handlers.OnFriendDel != nil {
				g.handlers.OnFriendDel(ev)
			}
		})
	case "C2C_MSG_REJECT":
		dispatchEvent(data, eventType, g.handlers.OnC2CMsgReject)
	case "C2C_MSG_RECEIVE":
		dispatchEvent(data, eventType, g.handlers.OnC2CMsgReceive)
	}
}
