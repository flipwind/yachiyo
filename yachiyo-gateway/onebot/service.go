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

	"github.com/gorilla/websocket"
)

var ylog = logger.New("Yachiyo.Adapter")
var upgrader = websocket.Upgrader{}
var channel *adapter.AdapterChannel
var conn *websocket.Conn

func handleReceive(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ylog.Error("Websocket upgrade error: %v", err)
		return
	}
	defer c.Close()

	conn = c

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			ylog.Error("Reading message error: %v", err)
			break
		}

		var messageEvent onebotModel.GroupMessageEvent
		err = json.Unmarshal(message, &messageEvent)
		if err != nil {
			ylog.Error("Received unsupported message: %v", err)
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
		// Handle sending action
		for msg := range channel.ToClient {
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

			conn.WriteJSON(req)
		}
	}()

	go func() {
		ylog.Success("Running onebot server on :16801 successfully.")
		if err := http.ListenAndServe(":16801", nil); err != nil {
			ylog.Error("Onebot Adapter running error: %v", err)
		}
	}()

}
