package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
)

type captureSender struct{ msgs []yws.ClientMessage }

func (c *captureSender) SendClientMessage(_ context.Context, msg yws.ClientMessage) error {
	c.msgs = append(c.msgs, msg)
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResp(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func newTestAPI(defaultMD bool, fn func(*http.Request) (*http.Response, error)) *qq.APIClient {
	hc := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "bots.qq.com" {
			return jsonResp(http.StatusOK, `{"access_token":"tok","expires_in":7200}`), nil
		}
		return fn(req)
	})}
	return qq.NewAPIClient(qq.NewTokenProvider("app", "sec", hc), qq.APIClientOptions{
		DefaultMarkdown: defaultMD, FallbackToPlainText: true, HTTPClient: hc,
	})
}

func TestFormatAuthor(t *testing.T) {
	if got := FormatAuthor("小红", "OPENID1"); got != "小红 [openid:OPENID1]" {
		t.Fatalf("got %q", got)
	}
	if got := FormatAuthor("", "OPENID1"); got != "user [openid:OPENID1]" {
		t.Fatalf("got %q", got)
	}
	if got := FormatAuthor("  ", ""); got != "user" {
		t.Fatalf("got %q", got)
	}
}

func TestInboundMentionReplacementAndDedup(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	st.UpdateMember("GRP1", store.CachedMember{MemberOpenID: "M_CACHED", Username: "阿明"})
	sender := &captureSender{}
	br := New(nil, st, sender)
	br.SetBotIdentity("BOT_ID", "八千代")

	ev := qq.GroupMessageEvent{
		ID: "MSG_1", GroupOpenID: "GRP1",
		Content:  "<@BOT_ID> 问一下 <@M_CACHED> 和 <@M_MENTION> 还有 <@M_UNKNOWN> <@all>",
		Author:   qq.GroupAuthor{MemberOpenID: "SENDER_1", Username: "小红"},
		Mentions: []qq.MentionUser{{MemberOpenID: "M_MENTION", Username: "大华"}},
	}
	br.HandleGroupMessage(context.Background(), ev, false)
	br.HandleGroupMessage(context.Background(), ev, true)

	if len(sender.msgs) != 1 {
		t.Fatalf("expected 1 forwarded message, got %d", len(sender.msgs))
	}
	got := sender.msgs[0]
	wantContent := "@八千代 问一下 @阿明 和 @大华 还有 @user [openid:M_UNKNOWN] @全体成员"
	if got.Author != "小红 [openid:SENDER_1]" || got.Path != "group/GRP1" || got.Message != wantContent {
		t.Fatalf("unexpected forwarded message: %+v", got)
	}
}

func TestOutboundMarkdownFallbackPreservesAtTag(t *testing.T) {
	var received []qq.SendGroupMessageRequest
	api := newTestAPI(true, func(req *http.Request) (*http.Response, error) {
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		received = append(received, body)
		if body.MsgType == qq.MsgTypeMarkdown {
			return jsonResp(http.StatusForbidden, `{"code":11292,"message":"markdown not allowed"}`), nil
		}
		return jsonResp(http.StatusOK, `{"id":"BOT_MSG_1"}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	br := New(api, st, &captureSender{})

	atMsg := `你好 <qqbot-at-user id="USER_1" />`
	br.HandleRuntimeMessage(context.Background(), yws.RuntimeMessage{Reply: true, Message: []string{atMsg}, Path: "group/GRP1"})

	if len(received) != 2 || received[0].MsgType != qq.MsgTypeMarkdown || received[1].MsgType != qq.MsgTypeText || received[1].Content != atMsg {
		t.Fatalf("unexpected fallback sequence: %+v", received)
	}
	if g, ok := st.GetGroup("GRP1"); !ok || g.LastBotMsgID != "BOT_MSG_1" {
		t.Fatalf("expected LastBotMsgID=BOT_MSG_1, got %+v", g)
	}
}

func TestC2CInboundAndOutboundWithPancakePacing(t *testing.T) {
	var paths []string
	var received []qq.SendGroupMessageRequest
	api := newTestAPI(true, func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.URL.Path)
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		received = append(received, body)
		return jsonResp(http.StatusOK, `{"id":"C2C_BOT_1"}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	sender := &captureSender{}
	br := New(api, st, sender)

	var delays []time.Duration
	br.sleepFn = func(_ context.Context, d time.Duration) bool { delays = append(delays, d); return true }

	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{
		ID: "C2C_IN_1", Content: "你好呀", Author: qq.C2CAuthor{UserOpenID: "USER_OPEN_1", Username: "小蓝"},
	})
	if len(sender.msgs) != 1 || sender.msgs[0].Path != "private/USER_OPEN_1" || sender.msgs[0].Author != "小蓝 [openid:USER_OPEN_1]" {
		t.Fatalf("unexpected C2C inbound: %+v", sender.msgs)
	}

	br.HandleRuntimeMessage(context.Background(), yws.RuntimeMessage{
		Reply: true, Message: []string{"第一条", "hi", strings.Repeat("字", 60)}, Path: "private/USER_OPEN_1",
	})
	if len(received) != 3 {
		t.Fatalf("expected 3 C2C outbound messages, got %d", len(received))
	}
	for i, p := range paths {
		if p != "/v2/users/USER_OPEN_1/messages" || received[i].MsgID != "C2C_IN_1" || received[i].MsgSeq != i+1 {
			t.Fatalf("req %d mismatch: path=%q msg_id=%q seq=%d", i, p, received[i].MsgID, received[i].MsgSeq)
		}
	}
	if len(delays) != 2 || delays[0] != 400*time.Millisecond || delays[1] != 5000*time.Millisecond || BubbleDelay("") != 300*time.Millisecond {
		t.Fatalf("unexpected delays: %v", delays)
	}
}

func TestC2CPassiveLimit4AndRejectProactiveAndEventTimeTTL(t *testing.T) {
	var sentReqs []qq.SendGroupMessageRequest
	api := newTestAPI(false, func(req *http.Request) (*http.Response, error) {
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		sentReqs = append(sentReqs, body)
		return jsonResp(http.StatusOK, `{"id":"OUT_OK"}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	br := New(api, st, &captureSender{})
	br.sleepFn = func(context.Context, time.Duration) bool { return true }

	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{
		ID: "OLD_C2C", Content: "old", Timestamp: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		Author: qq.C2CAuthor{UserOpenID: "U_OLD", Username: "Old"},
	})
	if _, _, ok := st.PeekPassiveAnchor("private", "U_OLD", ""); ok {
		t.Fatal("expected 2-hour-old event not to be usable as passive anchor")
	}

	st.SetC2CRejectProactive("U_LIMIT", true)
	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{
		ID: "FRESH_C2C", Content: "hi", Timestamp: time.Now().Format(time.RFC3339),
		Author: qq.C2CAuthor{UserOpenID: "U_LIMIT", Username: "User"},
	})
	br.HandleRuntimeMessage(context.Background(), yws.RuntimeMessage{
		Reply: true, Message: []string{"1", "2", "3", "4", "5"}, Path: "private/U_LIMIT",
	})
	if len(sentReqs) != 4 {
		t.Fatalf("expected exactly 4 passive C2C replies, got %d", len(sentReqs))
	}
	for i, r := range sentReqs {
		if r.MsgID != "FRESH_C2C" || r.MsgSeq != i+1 {
			t.Fatalf("bubble %d: msg_id=%q seq=%d", i, r.MsgID, r.MsgSeq)
		}
	}
}

func TestRateLimitDoesNotBurnPassiveQuota(t *testing.T) {
	calls := 0
	var seqs []int
	api := newTestAPI(false, func(req *http.Request) (*http.Response, error) {
		calls++
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		seqs = append(seqs, body.MsgSeq)
		if calls == 1 {
			return jsonResp(http.StatusTooManyRequests, `{"code":429,"message":"rate limited"}`), nil
		}
		return jsonResp(http.StatusOK, `{"id":"OK_1"}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	br := New(api, st, &captureSender{})
	br.sleepFn = func(context.Context, time.Duration) bool { return true }

	st.RecordInboundMessage("G_RL", store.CachedMessage{MessageID: "IN_1", Content: "hi", Timestamp: time.Now().Unix()}, nil)
	br.HandleRuntimeMessage(context.Background(), yws.RuntimeMessage{Reply: true, Message: []string{"a", "b"}, Path: "group/G_RL"})
	if len(seqs) != 2 || seqs[0] != 1 || seqs[1] != 1 {
		t.Fatalf("expected 429 not to advance seq (both seq=1), got %v", seqs)
	}
}

func TestPerSessionOutboundOrderingNoInterleave(t *testing.T) {
	var sentTexts []string
	api := newTestAPI(false, func(req *http.Request) (*http.Response, error) {
		var body qq.SendGroupMessageRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		sentTexts = append(sentTexts, body.Content)
		return jsonResp(http.StatusOK, `{"id":"OK"}`), nil
	})
	st := store.New(store.Options{RecentLimit: 100})
	br := New(api, st, &captureSender{})
	br.sleepFn = func(context.Context, time.Duration) bool { time.Sleep(15 * time.Millisecond); return true }

	br.EnqueueRuntimeMessage(context.Background(), yws.RuntimeMessage{Reply: true, Message: []string{"A1", "A2"}, Path: "group/G_ORD"})
	br.HandleRuntimeMessage(context.Background(), yws.RuntimeMessage{Reply: true, Message: []string{"B1"}, Path: "group/G_ORD"})
	if len(sentTexts) != 3 || sentTexts[0] != "A1" || sentTexts[1] != "A2" || sentTexts[2] != "B1" {
		t.Fatalf("expected strict FIFO order [A1 A2 B1], got %v", sentTexts)
	}
}

type flakySender struct {
	fail bool
	msgs []yws.ClientMessage
}

func (f *flakySender) SendClientMessage(_ context.Context, msg yws.ClientMessage) error {
	if f.fail {
		return io.ErrClosedPipe
	}
	f.msgs = append(f.msgs, msg)
	return nil
}

func TestPendingInboundFlushAndGroupReplyModeAndBotFilter(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	sender := &flakySender{fail: true}
	br := NewWithOptions(nil, st, sender, Options{GroupReplyMode: "at_only"})
	br.SetBotIdentity("BOT_1", "八千代")

	br.HandleGroupMessage(context.Background(), qq.GroupMessageEvent{
		ID: "M_PLAIN", GroupOpenID: "G1", Content: "普通聊天", Author: qq.GroupAuthor{MemberOpenID: "U1", Username: "小明"},
	}, false)
	if len(st.GetRecentMessages("G1", 10)) != 1 {
		t.Fatal("expected non-@ message to be cached")
	}

	br.HandleGroupMessage(context.Background(), qq.GroupMessageEvent{
		ID: "M_BOT", GroupOpenID: "G1", Content: "<@BOT_1> hello", Author: qq.GroupAuthor{MemberOpenID: "OTHER_BOT", Username: "Bot2", Bot: true},
	}, true)
	br.HandleGroupMessage(context.Background(), qq.GroupMessageEvent{
		ID: "M_AT", GroupOpenID: "G1", Content: "<@BOT_1> 帮我查一下", Author: qq.GroupAuthor{MemberOpenID: "U1", Username: "小明"},
	}, true)
	if len(sender.msgs) != 0 {
		t.Fatalf("expected 0 delivered while disconnected, got %d", len(sender.msgs))
	}

	sender.fail = false
	if flushed := br.FlushPending(context.Background()); flushed != 1 || len(sender.msgs) != 1 || !strings.Contains(sender.msgs[0].Message, "帮我查一下") {
		t.Fatalf("expected 1 flushed @ message, got flushed=%d msgs=%+v", flushed, sender.msgs)
	}
}

func TestIsYouMentionAndSeparatedCacheTriggerDedup(t *testing.T) {
	st := store.New(store.Options{RecentLimit: 100})
	sender := &captureSender{}
	br := NewWithOptions(nil, st, sender, Options{GroupReplyMode: "at_only"})

	br.HandleGroupMessage(context.Background(), qq.GroupMessageEvent{
		ID: "M_IS_YOU", GroupOpenID: "G1", Content: "你好",
		Author:   qq.GroupAuthor{MemberOpenID: "U1", Username: "小明"},
		Mentions: []qq.MentionUser{{MemberOpenID: "BOT_OPENID", IsYou: true}},
	}, false)
	if len(sender.msgs) != 1 {
		t.Fatalf("expected is_you=true to forward, got %d", len(sender.msgs))
	}

	ev := qq.GroupMessageEvent{
		ID: "M_DUAL", GroupOpenID: "G1", Content: "帮我看看", Author: qq.GroupAuthor{MemberOpenID: "U1", Username: "小明"},
	}
	br.HandleGroupMessage(context.Background(), ev, false)
	if len(sender.msgs) != 1 {
		t.Fatalf("non-@ first delivery should not forward yet, got %d", len(sender.msgs))
	}
	br.HandleGroupMessage(context.Background(), ev, true)
	br.HandleGroupMessage(context.Background(), ev, true)
	if len(sender.msgs) != 2 || len(st.GetRecentMessages("G1", 10)) != 2 {
		t.Fatalf("unexpected msgs=%d cached=%d", len(sender.msgs), len(st.GetRecentMessages("G1", 10)))
	}
}

func TestEnqueueRuntimeMessageNeverBlocksCaller(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	unblock := make(chan struct{})
	defer close(unblock)

	api := newTestAPI(false, func(req *http.Request) (*http.Response, error) {
		select {
		case <-unblock:
			return jsonResp(http.StatusOK, `{"id":"OK"}`), nil
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	})
	br := New(api, store.New(store.Options{RecentLimit: 100}), &captureSender{})

	start := time.Now()
	for i := 0; i < 200; i++ {
		br.EnqueueRuntimeMessage(ctx, yws.RuntimeMessage{Reply: true, Message: []string{"msg"}, Path: "group/G_BACKLOG"})
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("EnqueueRuntimeMessage blocked caller for %v", elapsed)
	}
}

type staleOnceSender struct {
	failOnce  bool
	connected bool
	msgs      []yws.ClientMessage
}

func (s *staleOnceSender) IsConnected() bool { return s.connected }
func (s *staleOnceSender) SendClientMessage(_ context.Context, msg yws.ClientMessage) error {
	if s.failOnce {
		s.failOnce = false
		return io.ErrClosedPipe
	}
	if !s.connected {
		return io.ErrClosedPipe
	}
	s.msgs = append(s.msgs, msg)
	return nil
}

func TestReconnectOrderAndStaleConnImmediateRetry(t *testing.T) {
	sender := &staleOnceSender{connected: false}
	br := New(nil, store.New(store.Options{RecentLimit: 100}), sender)

	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{ID: "C_1", Content: "first", Author: qq.C2CAuthor{UserOpenID: "U1", Username: "Alice"}})
	sender.connected = true
	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{ID: "C_2", Content: "second", Author: qq.C2CAuthor{UserOpenID: "U1", Username: "Alice"}})
	if len(sender.msgs) != 2 || sender.msgs[0].Message != "first" || sender.msgs[1].Message != "second" {
		t.Fatalf("expected strict order [first, second], got %+v", sender.msgs)
	}

	sender.failOnce = true
	br.HandleC2CMessage(context.Background(), qq.C2CMessageEvent{ID: "C_3", Content: "third", Author: qq.C2CAuthor{UserOpenID: "U1", Username: "Alice"}})
	if len(sender.msgs) != 3 || sender.msgs[2].Message != "third" {
		t.Fatalf("expected immediate flush on live connection for C_3, got %+v", sender.msgs)
	}
}

type ctxCheckSender struct {
	done      chan struct{}
	sawCtxErr error
	msgs      []yws.ClientMessage
}

func (c *ctxCheckSender) SendClientMessage(ctx context.Context, msg yws.ClientMessage) error {
	defer func() {
		select {
		case c.done <- struct{}{}:
		default:
		}
	}()
	if err := ctx.Err(); err != nil {
		c.sawCtxErr = err
		return err
	}
	c.msgs = append(c.msgs, msg)
	return nil
}

func TestCancelledQQSessionCtxDoesNotCancelWSJSONWrite(t *testing.T) {
	sender := &ctxCheckSender{done: make(chan struct{}, 1)}
	br := New(nil, store.New(store.Options{RecentLimit: 100}), sender)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	br.EnqueueC2CMessage(cancelledCtx, qq.C2CMessageEvent{ID: "C_CANCELLED", Content: "still forward", Author: qq.C2CAuthor{UserOpenID: "U1", Username: "Alice"}})
	select {
	case <-sender.done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for inbound message forward")
	}
	if sender.sawCtxErr != nil || len(sender.msgs) != 1 || sender.msgs[0].Message != "still forward" {
		t.Fatalf("unexpected result: err=%v msgs=%+v", sender.sawCtxErr, sender.msgs)
	}
}
