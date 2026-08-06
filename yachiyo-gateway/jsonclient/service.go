package jsonclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var ylog = logger.New("Yachiyo.JsonClient")
var channel *gateway.GatewayChannel

type JsonPack struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type MessagePack struct {
	Role      string `json:"role"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

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
				case *action.Status:
					content := JsonPack{
						Type: "status",
						Data: t.Content,
					}
					byteContent, err := json.Marshal(content)
					if err != nil {
						ylog.Error("Json marshal error: %v", err)
						return
					}

					if err := c.Write(ctx, websocket.MessageText, byteContent); err != nil {
						ylog.Error("Websocket writing error: %v", err)
						return
					}
				case *action.Message:
					content := JsonPack{
						Type: "message_delta",
						Data: MessagePack{
							Role: "assistant",
							Message: t.Content,
							Timestamp: t.Time,
						},
					}
					byteContent, err := json.Marshal(content)
					if err != nil {
						ylog.Error("Json marshal error: %v", err)
						return
					}
					
					if err := c.Write(ctx, websocket.MessageText, byteContent); err != nil {
						ylog.Error("Websocket writing error: %v", err)
						return
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

		var msgData JsonPack

		if err := json.Unmarshal(msg, &msgData); err != nil {
			ylog.Error("Unmarshal json {%v} error: %v", string(msg), err)
		}

		if msgData.Type == "send_message"{
			channel.ToServer <- &trigger.Message{
				Type:     "User",
				Author:   "user",
				Platform: "jsonclient",
				Time:     time.Now().Unix(),
				Content:  msgData.Data.(string),

				Address: address.Address{
					Content: "jsonclient://jsonclient",
				},
			}
		}
	}
}

type JsonClientService struct{}

func (s *JsonClientService) Listen(c *gateway.GatewayChannel, p int64) {
	channel = c

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", handleReceive)

	go func() {
		port := fmt.Sprintf(":%d", p)
		ylog.Success("Running client server on :%d successfully.", p)
		if err := http.ListenAndServe(port, mux); err != nil {
			ylog.Error("json client adapter running error: %v", err)
		}
	}()
}

func (s *JsonClientService) SchemeName() string {
	return "jsonclient"
}
