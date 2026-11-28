package yws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"uuid"

	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var (
	ylog            = logger.New("Yachiyo.QQ.WS")
	validClientName = regexp.MustCompile(`^[a-zA-Z0-9_]{1,16}$`)
)

type Handlers struct {
	OnConnected      func(ctx context.Context)
	OnRuntimeMessage func(ctx context.Context, msg RuntimeMessage)
	OnRuntimeError   func(ctx context.Context, errMsg RuntimeError)
	OnToolCall       func(ctx context.Context, call ToolCall) (bool, string)
}

type Options struct {
	GatewayURL        string
	ClientID          string
	ClientName        string
	HeartbeatInterval time.Duration
	Capabilities      []string
	Tools             []ToolItem
	Handlers          Handlers
}

type Client struct {
	opts       Options
	mu         sync.Mutex
	conn       *websocket.Conn
	registered bool
}

func NewClient(opts Options) (*Client, error) {
	parsed, err := uuid.Parse(opts.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client_id %q: %w", opts.ClientID, err)
	}
	opts.ClientID = parsed.String()
	if !validClientName.MatchString(opts.ClientName) {
		return nil, fmt.Errorf("invalid client_name %q: must match ^[a-zA-Z0-9_]{1,16}$", opts.ClientName)
	}
	if opts.GatewayURL == "" {
		return nil, errors.New("gateway_url cannot be empty")
	}
	if opts.HeartbeatInterval <= 0 || opts.HeartbeatInterval > 20*time.Second {
		opts.HeartbeatInterval = 15 * time.Second
	}
	return &Client{opts: opts}, nil
}

func (c *Client) ClientID() string { return c.opts.ClientID }

func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && c.registered
}

func (c *Client) Run(ctx context.Context) error {
	backoff := time.Second
	const maxBackoff = 30 * time.Second
	for {
		if ctx.Err() != nil {
			return nil
		}
		fatal, err := c.runOnce(ctx)
		if fatal {
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			ylog.Warn("Gateway connection closed: %v. Reconnecting in %s...", err, backoff)
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (c *Client) runOnce(ctx context.Context) (bool, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	conn, _, err := websocket.Dial(dialCtx, c.opts.GatewayURL, nil)
	cancel()
	if err != nil {
		return false, fmt.Errorf("dial %s: %w", c.opts.GatewayURL, err)
	}

	conn.SetReadLimit(16 << 20)
	c.setConn(conn, false)
	defer func() {
		if ctx.Err() != nil {
			c.sendOfflineBestEffort(conn)
		}
		c.clearConn(conn)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	if err := c.writeEnvelope(ctx, conn, "connection", "register", Register{
		ClientType: "IM", ClientName: c.opts.ClientName, ClientID: c.opts.ClientID,
	}); err != nil {
		return false, fmt.Errorf("send register: %w", err)
	}

	regCtx, regCancel := context.WithTimeout(ctx, 10*time.Second)
	_, raw, err := conn.Read(regCtx)
	regCancel()
	if err != nil {
		return false, fmt.Errorf("read register response: %w", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return false, fmt.Errorf("decode register response: %w", err)
	}
	logGatewayFrame("<-", env.Category, env.Type, env.Data)
	if env.Category != "connection" {
		return false, fmt.Errorf("unexpected register response category %q", env.Category)
	}
	switch env.Type {
	case "register_success":
		var okData RegisterSuccess
		_ = json.Unmarshal(env.Data, &okData)
		ylog.Success("Connected to Yachiyo runtime [%s %s] as @%s (%s)",
			okData.RuntimeName, okData.RuntimeVersion, c.opts.ClientName, c.opts.ClientID)
	case "register_error":
		var errData RegisterError
		_ = json.Unmarshal(env.Data, &errData)
		fatal := errData.ErrorType == "client_id_invalid" || errData.ErrorType == "client_name_invalid" ||
			errData.ErrorType == "client_info_error" || errData.ErrorType == "client_conflict"
		return fatal, fmt.Errorf("register rejected: %s", errData.ErrorType)
	default:
		return false, fmt.Errorf("unexpected register response type %q", env.Type)
	}

	c.setConn(conn, true)
	if err := c.writeEnvelope(ctx, conn, "tool", "update_tools", UpdateTools{
		Capability: c.opts.Capabilities, Tools: c.opts.Tools,
	}); err != nil {
		return false, fmt.Errorf("send update_tools: %w", err)
	}

	sessionCtx, sessionCancel := context.WithCancel(ctx)
	defer sessionCancel()
	if c.opts.Handlers.OnConnected != nil {
		go c.opts.Handlers.OnConnected(sessionCtx)
	}
	go c.heartbeatLoop(sessionCtx, conn)

	for {
		_, data, err := conn.Read(sessionCtx)
		if err != nil {
			return false, err
		}
		c.handleIncoming(sessionCtx, conn, data)
	}
}

func (c *Client) heartbeatLoop(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(c.opts.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.writeEnvelope(hbCtx, conn, "connection", "heartbeat", HeartBeat{})
			cancel()
			if err != nil {
				ylog.Warn("Heartbeat send failed: %v", err)
				return
			}
		}
	}
}

func (c *Client) handleIncoming(ctx context.Context, conn *websocket.Conn, raw []byte) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		ylog.Error("Decode wsjson envelope error: %v", err)
		return
	}
	logGatewayFrame("<-", env.Category, env.Type, env.Data)

	switch env.Category + "/" + env.Type {
	case "connection/heartbeat_respond":
		ylog.Debug("Heartbeat acknowledged")
	case "connection/register_error":
		var errData RegisterError
		_ = json.Unmarshal(env.Data, &errData)
		ylog.Error("Gateway connection error: %s", errData.ErrorType)
	case "interaction/runtime_message":
		var msg RuntimeMessage
		if err := json.Unmarshal(env.Data, &msg); err != nil {
			ylog.Error("Decode runtime_message error: %v", err)
		} else if c.opts.Handlers.OnRuntimeMessage != nil {
			c.opts.Handlers.OnRuntimeMessage(ctx, msg)
		}
	case "interaction/runtime_error":
		var errMsg RuntimeError
		if err := json.Unmarshal(env.Data, &errMsg); err != nil {
			ylog.Error("Decode runtime_error error: %v", err)
		} else if c.opts.Handlers.OnRuntimeError != nil {
			go c.opts.Handlers.OnRuntimeError(ctx, errMsg)
		} else {
			ylog.Warn("Runtime error [%s] (path=%s): %s", errMsg.Code, shortPath(errMsg.Path), errMsg.Message)
		}
	case "tool/update_tools_feedback":
		var fb UpdateToolsFeedback
		if err := json.Unmarshal(env.Data, &fb); err != nil {
			ylog.Error("Decode update_tools_feedback error: %v", err)
		} else if fb.Status == "success" {
			ylog.Success("Tools and capabilities registered successfully (%d tools)", len(c.opts.Tools))
		} else {
			ylog.Warn("update_tools partial/error: reject_capability=%v reject_tools=%v", fb.RejectCapability, fb.RejectTools)
		}
	case "tool/tool_call":
		var call ToolCall
		if err := json.Unmarshal(env.Data, &call); err != nil {
			ylog.Error("Decode tool_call error: %v", err)
		} else {
			go c.execToolCall(ctx, conn, call)
		}
	}
}

func (c *Client) execToolCall(ctx context.Context, conn *websocket.Conn, call ToolCall) {
	ok, result := false, "tool handler not configured"
	if c.opts.Handlers.OnToolCall != nil {
		ok, result = c.opts.Handlers.OnToolCall(ctx, call)
	}
	sendCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.writeEnvelope(sendCtx, conn, "tool", "tool_call_result", ToolCallResult{
		RequestID: call.RequestID, OK: ok, Result: result,
	}); err != nil {
		ylog.Warn("Send tool_call_result (%s) failed: %v", call.ToolName, err)
	}
}

func (c *Client) SendClientMessage(ctx context.Context, msg ClientMessage) error {
	c.mu.Lock()
	conn, reg := c.conn, c.registered
	c.mu.Unlock()
	if conn == nil || !reg {
		return errors.New("yachiyo gateway not connected")
	}
	return c.writeEnvelope(ctx, conn, "interaction", "client_message", msg)
}

func (c *Client) writeEnvelope(ctx context.Context, conn *websocket.Conn, category, typ string, payload any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, raw, err := marshalEnvelope(category, typ, payload)
	if err != nil {
		return err
	}
	c.mu.Lock()
	err = conn.Write(ctx, websocket.MessageText, raw)
	c.mu.Unlock()
	if err != nil {
		ylog.Warn("-> [%s/%s] send failed (bytes=%d): %v", category, typ, len(data), err)
		ylog.Debug("-> [%s/%s] failed raw=%s", category, typ, string(data))
		return err
	}
	logGatewayFrame("->", category, typ, data)
	return nil
}

func marshalEnvelope(category, typ string, payload any) (json.RawMessage, []byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}
	raw, err := json.Marshal(Envelope{Category: category, Type: typ, Data: data})
	return data, raw, err
}

func logGatewayFrame(direction, category, typ string, data json.RawMessage) {
	if category == "connection" && (typ == "heartbeat" || typ == "heartbeat_respond") {
		ylog.Debug("%s [%s/%s]", direction, category, typ)
		return
	}
	if category == "interaction" && typ == "client_message" {
		ylog.Debug("%s [%s/%s] raw=%s", direction, category, typ, string(data))
		return
	}
	if summary := summarizeFrame(category+"/"+typ, data); summary == "" {
		ylog.Info("%s [%s/%s]", direction, category, typ)
	} else {
		ylog.Info("%s [%s/%s] %s", direction, category, typ, summary)
	}
	ylog.Debug("%s [%s/%s] raw=%s", direction, category, typ, string(data))
}

func summarizeFrame(key string, data json.RawMessage) string {
	switch key {
	case "connection/register":
		var v Register
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("type=%s name=%s id=%s", v.ClientType, v.ClientName, shortID(v.ClientID))
		}
	case "connection/register_success":
		var v RegisterSuccess
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("runtime=%s %s", v.RuntimeName, v.RuntimeVersion)
		}
	case "connection/register_error":
		var v RegisterError
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("error=%s", v.ErrorType)
		}
	case "connection/offline":
		return ""
	case "tool/update_tools":
		var v UpdateTools
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("capabilities=%v tools=%d", v.Capability, len(v.Tools))
		}
	case "tool/update_tools_feedback":
		var v UpdateToolsFeedback
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("status=%s reject_cap=%d reject_tools=%d", v.Status, len(v.RejectCapability), len(v.RejectTools))
		}
	case "tool/tool_call":
		var v ToolCall
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("name=%s req_id=%s payload_bytes=%d", v.ToolName, shortID(v.RequestID), len(v.Payload))
		}
	case "tool/tool_call_result":
		var v ToolCallResult
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("req_id=%s ok=%t result_bytes=%d", shortID(v.RequestID), v.OK, len(v.Result))
		}
	case "interaction/runtime_message":
		var v RuntimeMessage
		if json.Unmarshal(data, &v) == nil {
			prefix := ""
			if v.IsInitiative {
				prefix = "[initiative] "
			}
			return fmt.Sprintf("[%s] %sbubbles=%d", shortPath(v.Path), prefix, len(v.Message))
		}
	case "interaction/runtime_error":
		var v RuntimeError
		if json.Unmarshal(data, &v) == nil {
			return fmt.Sprintf("[%s] code=%s msg=%q", shortPath(v.Path), v.Code, clipPreview(v.Message, 80))
		}
	}
	return fmt.Sprintf("bytes=%d", len(data))
}

func shortPath(path string) string {
	if idx := strings.IndexByte(path, '/'); idx >= 0 && idx < len(path)-1 {
		return path[:idx+1] + shortID(path[idx+1:])
	}
	return shortID(path)
}

func shortID(id string) string {
	if len(id) <= 10 {
		return id
	}
	return id[:4] + ".." + id[len(id)-4:]
}

func clipPreview(s string, maxRunes int) string {
	s = strings.NewReplacer("\r", "", "\n", " ").Replace(s)
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes]) + "..."
}

func (c *Client) setConn(conn *websocket.Conn, registered bool) {
	c.mu.Lock()
	c.conn, c.registered = conn, registered
	c.mu.Unlock()
}

func (c *Client) clearConn(conn *websocket.Conn) {
	c.mu.Lock()
	if c.conn == conn {
		c.conn, c.registered = nil, false
	}
	c.mu.Unlock()
}

func (c *Client) sendOfflineBestEffort(conn *websocket.Conn) {
	if !c.IsConnected() {
		return
	}
	_, raw, err := marshalEnvelope("connection", "offline", Offline{})
	if err != nil {
		return
	}
	offCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	c.mu.Lock()
	err = conn.Write(offCtx, websocket.MessageText, raw)
	c.mu.Unlock()
	if err == nil {
		ylog.Debug("-> [connection/offline]")
	}
}
