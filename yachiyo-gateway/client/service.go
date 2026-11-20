package client

import (
	"fmt"
	"net/http"
	"time"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var ylog = logger.New("Yachiyo.Client")

func (s *ClientService) handleReceive(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		ylog.Error("Websocket upgrade error: %v", err)
		return
	}
	defer c.CloseNow()

	ctx := r.Context()

	go func() {
		// Sending
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-s.out:
				switch t := msg.(type) {
				case *action.AssistantMessage:
					for _, message := range t.Content {
						if err := c.Write(ctx, websocket.MessageText, []byte(message)); err != nil {
							ylog.Error("Websocket writing error: %v", err)
							return
						}
					}
				}
			}
		}
	}()

	for {
		// Reading
		_, msg, err := c.Read(ctx)
		if err != nil {
			ylog.Error("Reading message error: %v", err)
			return
		}

		s.submit(&trigger.UserMessage{
			Author:    "user",
			Platform:  "CLI",
			Timestamp: time.Now(),
			Content:   string(msg),

			Address: address.New("client://cli"),
		})
	}
}

type ClientService struct {
	submit func(trig trigger.Trigger) //exhaustruct:optional
	out    chan action.Action
}

func New() *ClientService {
	return &ClientService{
		out: make(chan action.Action),
	}
}

func (s *ClientService) Listen(port int64, submit func(trig trigger.Trigger)) {
	s.submit = submit
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", s.handleReceive)

	go func() {
		addr := fmt.Sprintf(":%d", port)
		ylog.Success("Running client server on :%d successfully.", port)
		if err := http.ListenAndServe(addr, mux); err != nil {
			ylog.Error("Onebot Adapter running error: %v", err)
		}
	}()
}

func (s *ClientService) SchemeName() string {
	return "client"
}

func (s *ClientService) Send(a action.Action) {
	select {
	case s.out <- a:
	case <-time.After(time.Second):
		ylog.Error("%s send timeout, drop action <%v>", s.SchemeName(), a)
	}
}
