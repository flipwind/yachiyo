package jsonclient

import (
	"context"
	"encoding/json"
	"time"
	"yachiyo/yachiyo-gateway/jsonclient/model"

	"github.com/coder/websocket"
)

type Client struct {
	Type string
	Name string
	ID   string

	LastHeartbeatTime time.Time

	conn     *websocket.Conn
	sendChan chan model.Envelope
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:              conn,
		LastHeartbeatTime: time.Now(),
	}
}

func (c *Client) Run(unregisterHandler func(c *Client), dataHandler func(c *Client, data []byte)) {
	go c.readListen(unregisterHandler, dataHandler)
	go c.writeListen()
}

func (c *Client) readListen(unregisterHandler func(c *Client), dataHandler func(c *Client, data []byte)) {
	defer func() {
		unregisterHandler(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, data, err := c.conn.Read(
			context.Background(),
		)

		if err != nil {
			ylog.Error("Reading error: %v", err)
			return
		}

		dataHandler(c, data)
	}
}

func (c *Client) writeListen() {
	for data := range c.sendChan {
		dataJson, err := json.Marshal(data)
		if err != nil {
			ylog.Error("Data marshal to json error: %v", err)
			return
		}

		c.conn.Write(
			context.Background(),
			websocket.MessageText,
			dataJson,
		)
	}
}

func (c *Client) send(category string, contentType string, d model.DataPack) {
	dataJson, err := json.Marshal(d)
	if err != nil {
		ylog.Error("DataPack marshal to json error: %v", err)
		return
	}

	c.sendChan <- model.Envelope{
		Category: category,
		Type: contentType,
		Data: dataJson,
	}
}
