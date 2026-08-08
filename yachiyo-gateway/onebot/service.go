package onebot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/onebot/model"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/address"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var ylog = logger.New("Yachiyo.Onebot")
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
					u, err := url.Parse(t.Address.Content)
					if err != nil {
						ylog.Error("Address url parse error: %v", err)
						ylog.Info("Message <%v> is ignored.", msg)
						continue
					}

					req := map[string]any{
						"action": "send_group_msg",
						"params": map[string]any{
							"group_id":    strings.TrimPrefix(u.Path, "/"),
							"message":     t.Content,
							"auto_escape": true,
						},
					}

					jreq, _ := json.Marshal(req)

					if err := c.Write(ctx, websocket.MessageText, jreq); err != nil {
						ylog.Error("Websocket writing error: %v", err)
						return
					}
				}

			}
		}
	}()

	for {
		// Reading
		var messageEvent onebotModel.GroupMessageEvent
		if err := wsjson.Read(ctx, c, &messageEvent); err != nil {
			ylog.Error("Reading message error: %v", err)
			return
		}

		if messageEvent.PostType == "message" {
			// TODO：rich message supporter
			channel.ToServer <- &trigger.Message{
				Type:     "user",
				Author:   messageEvent.Sender.Nickname,
				Platform: "QQ",
				Time:     time.Now().Unix(),
				Content:  messageEvent.RawMessage,

				Address: address.Address{
					Content: fmt.Sprintf("onebot://group/%v", messageEvent.GroupID),
				},
			}
		}
	}
}

type OnebotService struct {}

func (s *OnebotService) Listen(c *gateway.GatewayChannel, p int64) {
	channel = c
	http.HandleFunc("/ws/onebot", handleReceive)

	go func() {
		port := fmt.Sprintf(":%d", p)
		ylog.Success("Running onebot server on :%d successfully.", p)
		if err := http.ListenAndServe(port, nil); err != nil {
			ylog.Error("Onebot Adapter running error: %v", err)
		}
	}()
}

func (s *OnebotService) SchemeName() string {
	return "onebot"
}
