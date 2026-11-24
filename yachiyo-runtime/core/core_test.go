package core_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store/sqlite3"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/endpoint"
	"yachiyo/yachiyo-shared/trigger"
)

type fakeLLM struct {
	called bool
}

func (l *fakeLLM) Generate(ctx context.Context, req llm.Request) (llm.Response, error) {
	if !l.called {
		l.called = true
		return llm.Response{
			FinishReason: llm.FinishToolCall,
			Content: "",
			ToolCalls: []llm.ToolCall{
				{
					ID:        "call_1",
					Name:      "commit_message",
					Arguments: `{"address":"test://test", "messages":["received+hello"]}`,
				},
			},
		}, nil
	}

	return llm.Response{
		FinishReason: llm.FinishStop,
		Content:      `{"thoughts":"ok","note":"-1"}`,
		ToolCalls:    make([]llm.ToolCall, 0),
	}, nil
}

type fakeHub struct {
	outbox chan action.Action
}

func newFakeHub(outbox chan action.Action) *fakeHub {
	return &fakeHub{
		outbox: outbox,
	}
}

func (h *fakeHub) Deliver(a action.Action) {
	h.outbox <- a
}
func (h *fakeHub) AvailableAddresses(cap string) ([]address.Address, error) {
	return make([]address.Address, 0), nil
}
func (h *fakeHub) ClientTools() []endpoint.RemoteTool {
	return make([]endpoint.RemoteTool, 0)
}
func (h *fakeHub) SendMessage(a action.AssistantMessage) error {
	h.outbox <- &a
	return nil
}
func (h *fakeHub) UseTool(ctx context.Context, addr address.Address, reqID, toolName, payload string) (string, error) {
	return "", nil
}

func fakeNewCore(t *testing.T) (*core.Core, chan action.Action) {
	t.Helper()

	fatal := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	energy, err := state.NewEnergy("1")
	fatal(err)

	st, err := sqlite3.Open(filepath.Join(t.TempDir(), "test.db"))
	fatal(err)

	t.Cleanup(func() { st.Close() })
	var outbox chan action.Action = make(chan action.Action)

	hub := newFakeHub(outbox)

	c, err := core.New(core.Options{
		SystemPrompt:   "system",
		CompressPrompt: "compress",
		Personality:    state.BigFive{},
		Energy:         energy,
	}, core.Deps{
		Store: st,
		LLM:   &fakeLLM{},
		Hub:   hub,
	})
	fatal(err)

	return c, outbox
}

func TestUserMessageReply(t *testing.T) {
	c, outbox := fakeNewCore(t)
	timer := time.NewTimer(time.Second * 3)

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { c.Run(ctx); close(done) }()
	defer func() { cancel(); <-done }()

	c.Dispatch(&trigger.UserMessage{
		Author:    "author",
		Platform:  "platform",
		Content:   "content",
		Timestamp: time.Now(),
		Address:   address.New("test://test"),
	})

	select {
	case result := <-outbox:
		r, ok := result.(*action.AssistantMessage)
		if !ok {
			t.Fatalf("want action.AssistantMessage, got %T: %v", result, result)
		}
		if !slices.Equal(r.Content, []string{"received+hello"}) {
			t.Fatalf("Reply doesn't match sended message")
		}
	case <-timer.C:
		t.Fatal("timeout")
	}
}
