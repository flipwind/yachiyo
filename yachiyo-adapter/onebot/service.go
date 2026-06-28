package onebot

import (
	"net/http"
	"yachiyo/yachiyo-utils/logger"
	"github.com/gorilla/websocket"
)

var log = logger.New("Yachiyo.Adapter")

var upgrader = websocket.Upgrader{}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Websocket upgrade error: %v", err)
		return
	}
	defer c.Close()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Error("Reading message error: %v", err)
			break
		}
	}
}

func Service() {
	http.HandleFunc("/ws/onebot", handleWebsocket)

	log.Success("Running onebot server on :16801 successfully.")
	if err := http.ListenAndServe(":16801", nil); err != nil {
		log.Error("Onebot Adapter running error: %v", err)
	}
}