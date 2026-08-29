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
		sendChan:          make(chan model.Envelope, 64),
		LastHeartbeatTime: time.Now(),
	}
}

func (c *Client) Run(unregisterHandler func(c *Client), dataHandler func(c *Client, data []byte)) {
	go c.readListen(unregisterHandler, dataHandler)
	go c.writeListen(unregisterHandler)
}

func (c *Client) readListen(unregisterHandler func(c *Client), dataHandler func(c *Client, data []byte)) {
	defer func() {
		unregisterHandler(c)
	}()

	for {
		_, data, err := c.conn.Read(
			context.Background(),
		)

		if err != nil {
			ylog.Error("Reading error: %v", err)
			break
		}

		dataHandler(c, data)
	}
}

func (c *Client) writeListen(unregisterHandler func(c *Client)) {
	defer func() {
		unregisterHandler(c)
	}()

	for data := range c.sendChan {
		dataJson, err := json.Marshal(data)
		if err != nil {
			ylog.Error("Data marshal to json error: %v", err)
			break
		}

		err = c.conn.Write(
			context.Background(),
			websocket.MessageText,
			dataJson,
		)

		if err != nil {
			ylog.Error("Can't write to connection: %s", c.ID)
			unregisterHandler(c)
			break
		}
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
		Type:     contentType,
		Data:     dataJson,
	}
}
