package client

import (
	"net/http"
	"time"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var ylog = logger.New("Yachiyo.Client")
var channel *gateway.GatewayChannel

func handleReceive(w http.ResponseWriter, r *http.Request) {
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
			case msg := <-channel.ToClient:
				switch t := msg.(type) {
				case *action.Message:
					if err != nil {
						ylog.Error("Address url parse error: %v", err)
					}
					c.Write(ctx, websocket.MessageText, []byte(t.Content))
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

		channel.ToServer <- &trigger.Message{
			Type:     "User",
			Author:   "flipwind",
			Platform: "CLI",
			Time:     time.Now().Unix(),
			Content:  string(msg),

			Address: trigger.Address{
				Content: "client://cli",
			},
		}
	}
}

type ClientService struct{}

func (s *ClientService) Listen(c *gateway.GatewayChannel) {
	channel = c
	http.HandleFunc("/ws/", handleReceive)

	go func() {
		ylog.Success("Running client server on :16802 successfully.")
		if err := http.ListenAndServe(":16802", nil); err != nil {
			ylog.Error("Onebot Adapter running error: %v", err)
		}
	}()
}

func (s *ClientService) SchemeName() string {
	return "client"
}
