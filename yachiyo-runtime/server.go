package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/client"
	"yachiyo/yachiyo-gateway/onebot"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

func main() {
	ylog.Info("Initializing Yachiyo server...")

	core := core.New()

	ylog.Success("Successfully initialize Yachiyo server.")

	go serviveChannel(core, onebot.Service)
	go serviveChannel(core, client.Service)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
}

func serviveChannel(c *core.Core, gw func(ch *gateway.GatewayChannel)) {
	gatewayChannel := gateway.NewGatewayChannel()
	gw(gatewayChannel)

	for msg := range gatewayChannel.ToServer {
		ylog.Debug("Processing [%s]", msg.Content)
		core_copy := *c
		result := c.Process(&msg)

		fmt.Printf("Yachiyo > %s\n", result)

		fmt.Printf("\nDEBUG\nFORMAL:")
		fmt.Println(core_copy.Emotion.String())
		fmt.Println(core_copy.State.Prompt())
		fmt.Printf("\nLATER:\n")
		fmt.Println(c.Emotion.String())
		fmt.Println(c.State.Prompt())
		fmt.Println(c.Determination)

		fmt.Println("Session Context: " + c.Note)

		msg.Content = result

		gatewayChannel.ToClient <- msg
	}
}
