package jsonclient

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

type Client struct {
	Type string
	Name string
	ID   string

	LastHeartbeatTime time.Time

	conn *websocket.Conn
	send chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
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
	for data := range c.send {
		c.conn.Write(
			context.Background(),
			websocket.MessageText,
			data,
		)
	}
}
