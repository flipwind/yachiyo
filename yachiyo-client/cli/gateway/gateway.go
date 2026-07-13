package gateway

import (
	"context"
	"github.com/coder/websocket"
	"time"
	"yachiyo/yachiyo-client/cli/event"
	"yachiyo/yachiyo-util/logger"
)

type CliChannel struct {
	ToClient chan event.Message
	ToServer chan event.Message
}

type Gateway struct {
	channel *CliChannel
}

var ylog = logger.New("Yachiyo.CLI")

func NewCliChannel() *CliChannel {
	return &CliChannel{
		ToClient: make(chan event.Message),
		ToServer: make(chan event.Message),
	}
}

func NewGateway(cc *CliChannel) *Gateway {
	return &Gateway{
		channel: cc,
	}
}

func (g *Gateway) Listen() {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	conn, _, err := websocket.Dial(dialCtx, "ws://localhost:16802/ws", nil)
	dialCancel()
	if err != nil {
		ylog.Error("Dialing error: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ylog.Success("Yachiyo CLI loading successfully.")

	// Sending
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-g.channel.ToServer:
				if err != nil {
					ylog.Error("Address url parse error: %v", err)
				}
				conn.Write(ctx, websocket.MessageText, []byte(msg.Content))
			}
		}
	}()

	// Reading
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			ylog.Error("Reading message error: %v", err)
			return
		}
		g.channel.ToClient <- event.Message{
			Content: string(msg),
		}
	}
}
