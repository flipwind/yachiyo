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
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var ylog = logger.New("Yachiyo.Adapter")
var channel *adapter.AdapterChannel

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
				u, err := url.Parse(msg.Address)
				if err != nil {
					ylog.Error("Address url parse error: %v", err)
				}

				req := map[string]any{
					"action": "send_group_msg",
					"params": map[string]any{
						"group_id":    strings.TrimPrefix(u.Path, "/"),
						"message":     msg.Content,
						"auto_escape": true,
					},
				}

				jreq, _ := json.Marshal(req)

				c.Write(ctx, websocket.MessageText, jreq)

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
			channel.ToServer <- trigger.Message{
				Type:     "User",
				Author:   messageEvent.Sender.Nickname,
				Platform: "QQ",
				Time:     time.Now().Unix(),
				Content:  messageEvent.RawMessage,

				Address: fmt.Sprintf("onebot://group/%v", messageEvent.GroupID),
			}
		}
	}
}

func Service(c *adapter.AdapterChannel) {
	channel = c
	http.HandleFunc("/ws/onebot", handleReceive)

	go func() {
		ylog.Success("Running onebot server on :16801 successfully.")
		if err := http.ListenAndServe(":16801", nil); err != nil {
			ylog.Error("Onebot Adapter running error: %v", err)
		}
	}()
}
