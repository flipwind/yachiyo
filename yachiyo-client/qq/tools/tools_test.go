package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
)

const testClientID = "01950000-0000-7000-8000-000000000001"

var validToolName = regexp.MustCompile(`^[a-zA-Z0-9_]{1,32}$`)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResp(code int, body string) *http.Response {
	return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(body))}
}

func newTestAPI(fn func(*http.Request) (*http.Response, error)) *qq.APIClient {
	hc := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return jsonResp(http.StatusOK, `{"access_token":"tok","expires_in":7200}`), nil
		}
		return fn(req)
	})}
	return qq.NewAPIClient(qq.NewTokenProvider("app", "sec", hc), qq.APIClientOptions{DefaultMarkdown: false, HTTPClient: hc})
}

func TestToolSpecsValid(t *testing.T) {
	reg := NewRegistry(testClientID, nil, store.New(store.Options{RecentLimit: 100}))
	specs := reg.Specs()
	if len(specs) != 7 {
		t.Fatalf("expected 7 tools, got %d", len(specs))
	}
	seen := make(map[string]bool)
	for _, s := range specs {
		if !validToolName.MatchString(s.Name) || seen[s.Name] || !json.Valid(s.Parameters) {
			t.Fatalf("invalid/duplicate spec: %s", s.Name)
		}
		seen[s.Name] = true
	}
}

func TestListGroupsAndRecentMessages100(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	for i := 0; i < 120; i++ {
		st.RecordInboundMessage("GRP_ABC", store.CachedMessage{
			MessageID: fmt.Sprintf("M_%d", i), MemberOpenID: "U1",
			Author: "alice [openid:U1]", Content: fmt.Sprintf("msg %d", i), Timestamp: int64(1000 + i),
		}, &store.CachedMember{MemberOpenID: "U1", Username: "alice"})
	}
	reg := NewRegistry(testClientID, nil, st)

	ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_list_groups", Payload: json.RawMessage(`{}`)})
	if !ok || !strings.Contains(res, `"path":"group/GRP_ABC"`) || !strings.Contains(res, `"address":"wsjson://`+testClientID+`/group/GRP_ABC"`) {
		t.Fatalf("qq_list_groups unexpected: ok=%t res=%s", ok, res)
	}

	ok, res = reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_get_recent_messages", Payload: json.RawMessage(`{"group_openid":"GRP_ABC","limit":100}`)})
	if !ok {
		t.Fatalf("qq_get_recent_messages failed: %s", res)
	}
	var parsed struct {
		Count    int                   `json:"count"`
		Messages []store.CachedMessage `json:"messages"`
	}
	if err := json.Unmarshal([]byte(res), &parsed); err != nil || parsed.Count != 100 || parsed.Messages[0].MessageID != "M_20" || parsed.Messages[99].MessageID != "M_119" {
		t.Fatalf("unexpected recent messages: err=%v parsed=%+v", err, parsed)
	}
}

func TestRecallBlocksOtherUsersMessage(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	st.RecordInboundMessage("GRP1", store.CachedMessage{MessageID: "OTHER_MSG", MemberOpenID: "U1", Author: "alice [openid:U1]", Content: "hi"}, nil)
	reg := NewRegistry(testClientID, nil, st)
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_recall_group_message", Payload: json.RawMessage(`{"group_openid":"GRP1","message_id":"OTHER_MSG"}`)}); ok {
		t.Fatalf("expected recall to fail, got ok=true: %s", res)
	}
}

func TestListPrivatesAndRecentPrivateMessages(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	st.RecordInboundC2C("U_PRIV_1", "IN_1", "小蓝", "小蓝 [openid:U_PRIV_1]", "你好", 1700000001)
	st.RecordBotC2C("U_PRIV_1", "OUT_1", "收到啦", 1700000002)
	reg := NewRegistry(testClientID, nil, st)

	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_list_privates", Payload: json.RawMessage(`{}`)}); !ok || !strings.Contains(res, `"path":"private/U_PRIV_1"`) {
		t.Fatalf("qq_list_privates unexpected: ok=%t res=%s", ok, res)
	}
	ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_get_recent_messages", Payload: json.RawMessage(`{"path":"private/U_PRIV_1","limit":10}`)})
	var parsed struct {
		Path     string                `json:"path"`
		Count    int                   `json:"count"`
		Messages []store.CachedMessage `json:"messages"`
	}
	if !ok || json.Unmarshal([]byte(res), &parsed) != nil || parsed.Path != "private/U_PRIV_1" || parsed.Count != 2 || parsed.Messages[0].IsSelf || !parsed.Messages[1].IsSelf {
		t.Fatalf("unexpected private messages: ok=%t res=%s", ok, res)
	}
}

func TestReplyUsesRefIdxAndRejectsMissingRefIdx(t *testing.T) {
	var sentRefID string
	api := newTestAPI(func(req *http.Request) (*http.Response, error) {
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		if body.MessageReference != nil {
			sentRefID = body.MessageReference.MessageID
		}
		return jsonResp(http.StatusOK, `{"id":"BOT_R1","ext_info":{"ref_idx":"REFIDX_BOT_R1"}}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	now := time.Now().Unix()
	st.RecordInboundMessage("G1", store.CachedMessage{MessageID: "PLAIN_ID_1", RefIdx: "REFIDX_USER_1", Content: "hello", Timestamp: now}, nil)
	st.RecordInboundMessage("G1", store.CachedMessage{MessageID: "OLD_NO_REF", Content: "legacy", Timestamp: now}, nil)
	reg := NewRegistry(testClientID, api, st)

	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_reply_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"PLAIN_ID_1","content":"reply"}`)}); !ok || sentRefID != "REFIDX_USER_1" {
		t.Fatalf("expected ref_idx reply, got ok=%t sentRefID=%q res=%s", ok, sentRefID, res)
	}
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_reply_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"OLD_NO_REF","content":"reply"}`)}); ok || !strings.Contains(res, "ref_idx") {
		t.Fatalf("expected ref_idx error, got ok=%t res=%s", ok, res)
	}
}

func TestRecallUpdatesCacheAndEnforcesOwnershipAndWindow(t *testing.T) {
	var recalledIDs []string
	api := newTestAPI(func(req *http.Request) (*http.Response, error) {
		parts := strings.Split(req.URL.Path, "/")
		recalledIDs = append(recalledIDs, parts[len(parts)-1])
		return jsonResp(http.StatusNoContent, ""), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	now := time.Now().Unix()
	st.RecordBotMessage("G1", "BOT_1", "first", now-10)
	st.RecordBotMessage("G1", "BOT_2", "second", now-5)
	st.RecordBotMessage("G1", "BOT_OLD", "old", now-300)
	reg := NewRegistry(testClientID, api, st)

	if ok, _ := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_recall_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"UNKNOWN_ID"}`)}); ok {
		t.Fatal("expected uncached ID to be rejected")
	}
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_recall_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"BOT_OLD"}`)}); ok || !strings.Contains(res, "2-minute") {
		t.Fatalf("expected 2-minute window error, got ok=%t res=%s", ok, res)
	}
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_recall_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"BOT_2"}`)}); !ok {
		t.Fatalf("recall BOT_2 failed: %s", res)
	}
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_recall_group_message", Payload: json.RawMessage(`{"group_openid":"G1","message_id":"BOT_2"}`)}); ok || !strings.Contains(res, "already been recalled") {
		t.Fatalf("expected already recalled error, got ok=%t res=%s", ok, res)
	}
	if len(recalledIDs) != 1 || recalledIDs[0] != "BOT_2" {
		t.Fatalf("unexpected recalledIDs: %v", recalledIDs)
	}
}

func TestGetGroupInfoPartialFailureAndMemberPermissionError(t *testing.T) {
	api := newTestAPI(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(req.URL.Path, "/info"):
			return jsonResp(http.StatusOK, `{"group_openid":"G1","group_name":"TestGroup"}`), nil
		case strings.HasSuffix(req.URL.Path, "/bot_state"):
			return jsonResp(http.StatusNotFound, `{"code":11001,"err_code":40011002,"message":"not found"}`), nil
		case strings.Contains(req.URL.Path, "/members/"):
			return jsonResp(http.StatusBadRequest, `{"code":11253,"err_code":40012010,"message":"no permission"}`), nil
		}
		return jsonResp(http.StatusNotFound, `{}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	st.UpdateMember("G1", store.CachedMember{MemberOpenID: "M1", Username: "Alice"})
	reg := NewRegistry(testClientID, api, st)

	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_get_group_info", Payload: json.RawMessage(`{"group_openid":"G1"}`)}); !ok || !strings.Contains(res, `"source":"api_partial"`) || !strings.Contains(res, `"bot_state_available":false`) {
		t.Fatalf("expected api_partial and bot_state_available=false, got ok=%t res=%s", ok, res)
	}
	if ok, res := reg.Execute(context.Background(), yws.ToolCall{ToolName: "qq_get_member_info", Payload: json.RawMessage(`{"group_openid":"G1","member_openid":"M1"}`)}); !ok || !strings.Contains(res, `"api_error_type":"permission_denied"`) {
		t.Fatalf("expected permission_denied fallback, got ok=%t res=%s", ok, res)
	}
}
