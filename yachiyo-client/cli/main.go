package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
	"yachiyo/yachiyo-util/logger"

	"github.com/coder/websocket"
)

var ylog = logger.New("Yachiyo.CLI")

func main() {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	c, _, err := websocket.Dial(dialCtx, "ws://localhost:16802/ws", nil)
	dialCancel()
	if err != nil {
		ylog.Error("Dialing error: %v", err)
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ylog.Success("Yachiyo CLI loading successfully.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Reading
		fmt.Print("\nUser > ")
		if !scanner.Scan() {
			return
		}

		input := scanner.Text()
		if err := c.Write(ctx, websocket.MessageText, []byte(input)); err != nil {
			ylog.Error("Write error: %v", err)
			return
		}

		_, reply, err := c.Read(ctx)
		if err != nil {
			ylog.Error("Reading error: %v", err)
			return
		}
		fmt.Printf("Yachiyo > %s\n", string(reply))

		if err := scanner.Err(); err != nil {
			ylog.Error("Scanner error: %v", err)
		}
	}
}
