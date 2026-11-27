package qq

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func tokenOKResp() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"tok","expires_in":7200}`)),
	}
}

func TestExtractMsgIdx(t *testing.T) {
	scene := &MessageScene{
		Source: "group",
		Ext:    []string{"ref_msg_idx=REFIDX_OLD", "msg_idx=REFIDX_TARGET"},
	}
	if got := ExtractMsgIdx(scene); got != "REFIDX_TARGET" {
		t.Fatalf("ExtractMsgIdx = %q, want REFIDX_TARGET", got)
	}
	if got := ExtractMsgIdx(nil); got != "" {
		t.Fatalf("ExtractMsgIdx(nil) = %q, want empty", got)
	}
}

func TestSandboxUsesUnifiedDomain(t *testing.T) {
	var seenHost string
	hc := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return tokenOKResp(), nil
		}
		seenHost = req.URL.Host
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"url":"wss://api.sgroup.qq.com/websocket"}`)),
		}, nil
	})}
	tokens := NewTokenProvider("app", "sec", hc)
	api := NewAPIClient(tokens, APIClientOptions{Sandbox: true, HTTPClient: hc})
	if _, err := api.GetGatewayURL(context.Background()); err != nil {
		t.Fatalf("GetGatewayURL failed: %v", err)
	}
	if seenHost != "api.bot.qq.com" {
		t.Fatalf("expected unified host api.bot.qq.com, got %q", seenHost)
	}
}

func TestMarkdownFallbackStructuredErrCodeAndNoAuthFallback(t *testing.T) {
	// 1. err_code=40034127 with non-English message must fall back to plain text.
	var attempts []int
	hc := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return tokenOKResp(), nil
		}
		var body SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		attempts = append(attempts, body.MsgType)
		if body.MsgType == MsgTypeMarkdown {
			h := make(http.Header)
			h.Set("X-Tps-Trace-Id", "TRACE-MD-1")
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     h,
				Body:       io.NopCloser(bytes.NewBufferString(`{"code":400,"err_code":40034127,"message":"不支持的消息类型"}`)),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"id":"MSG_OK","ext_info":{"ref_idx":"REFIDX_BOT_1"}}`)),
		}, nil
	})}
	tokens := NewTokenProvider("app", "sec", hc)
	api := NewAPIClient(tokens, APIClientOptions{DefaultMarkdown: true, FallbackToPlainText: true, HTTPClient: hc})
	resp, fallback, err := api.SendGroupMessage(context.Background(), "G1", "hi", SendOptions{})
	if err != nil {
		t.Fatalf("expected fallback to succeed, got err: %v", err)
	}
	if !fallback || resp.ID != "MSG_OK" || resp.RefIdx() != "REFIDX_BOT_1" {
		t.Fatalf("unexpected resp=%+v fallback=%t", resp, fallback)
	}
	if len(attempts) != 2 || attempts[0] != MsgTypeMarkdown || attempts[1] != MsgTypeText {
		t.Fatalf("unexpected attempts: %v", attempts)
	}

	// 2. HTTP 401 auth error must NOT fall back and must include status/code/err_code/trace_id.
	authAttempts := 0
	hc2 := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return tokenOKResp(), nil
		}
		authAttempts++
		h := make(http.Header)
		h.Set("X-Tps-Trace-Id", "TRACE-AUTH-1")
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     h,
			Body:       io.NopCloser(bytes.NewBufferString(`{"code":11244,"err_code":40011244,"message":"invalid token"}`)),
		}, nil
	})}
	api2 := NewAPIClient(NewTokenProvider("app", "sec", hc2), APIClientOptions{DefaultMarkdown: true, FallbackToPlainText: true, HTTPClient: hc2})
	_, _, err = api2.SendGroupMessage(context.Background(), "G1", "hi", SendOptions{})
	if err == nil {
		t.Fatal("expected auth error")
	}
	if authAttempts != 1 {
		t.Fatalf("401 should not trigger markdown fallback, got %d attempts", authAttempts)
	}
	errStr := err.Error()
	for _, want := range []string{"http=401", "code=11244", "err_code=40011244", "trace_id=TRACE-AUTH-1"} {
		if !strings.Contains(errStr, want) {
			t.Fatalf("error %q missing %q", errStr, want)
		}
	}
}

func TestSendResponseValidation(t *testing.T) {
	hc := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return tokenOKResp(), nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"id":[]}`)),
		}, nil
	})}
	api := NewAPIClient(NewTokenProvider("app", "sec", hc), APIClientOptions{DefaultMarkdown: false, HTTPClient: hc})
	if _, _, err := api.SendGroupMessage(context.Background(), "G1", "hi", SendOptions{}); err == nil {
		t.Fatal("expected error for malformed {\"id\":[]} response")
	}
}

func TestGetGroupBotStateUsesBotStatePath(t *testing.T) {
	var seenPath string
	hc := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return tokenOKResp(), nil
		}
		seenPath = req.URL.Path
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"is_muted":false}`)),
		}, nil
	})}
	api := NewAPIClient(NewTokenProvider("app", "sec", hc), APIClientOptions{HTTPClient: hc})
	if _, err := api.GetGroupBotState(context.Background(), "GRP1"); err != nil {
		t.Fatalf("GetGroupBotState failed: %v", err)
	}
	if seenPath != "/v2/groups/GRP1/bot_state" {
		t.Fatalf("expected /v2/groups/GRP1/bot_state, got %q", seenPath)
	}
}
