package onebot

import (
	"encoding/json"
	"net/http"
	"time"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-utils/logger"

	"github.com/gorilla/websocket"
)

var log = logger.New("Yachiyo.Adapter")
var upgrader = websocket.Upgrader{}
var channel *AdapterChannel
var conn *websocket.Conn

func handleReceive(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Websocket upgrade error: %v", err)
		return
	}
	defer c.Close()

	conn = c

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Error("Reading message error: %v", err)
			break
		}

		var messageEvent GroupMessageEvent
		err = json.Unmarshal(message, &messageEvent)
		if err != nil {
			log.Error("Received unsupported message: %v", err)
		}

		if messageEvent.PostType == "message" {
			// TODO：rich message supporter
			channel.ToServer <- event.UserMessageEvent{
				Message: event.Message{
					Type:    "User",
					Author:  messageEvent.Sender.Nickname,
					Source:  "QQ",
					Time:    time.Now().Unix(),
					Content: messageEvent.RawMessage,
					Payload: map[string]any{
						"group_id": messageEvent.GroupID,
					},
				},
			}
		}
	}
}

func Service(c *AdapterChannel) {
	channel = c
	http.HandleFunc("/ws/onebot", handleReceive)

	go func() {
		for event := range channel.ToClient {
			req := map[string]any{
				"action": "send_group_msg",
				"params": map[string]any{
					"group_id": event.GroupID,
					"message": event.Content,
					"auto_escape": true,
				},
			}

			conn.WriteJSON(req)
		}
	}()

	go func() {
		log.Success("Running onebot server on :16801 successfully.")
		if err := http.ListenAndServe(":16801", nil); err != nil {
			log.Error("Onebot Adapter running error: %v", err)
		}
	}()

}
