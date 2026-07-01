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

	var core_copy core.Core
	core := core.New()

	ylog.Success("Successfully initialize Yachiyo server.")

	gatewayChanOnebot := gateway.NewGatewayChannel()
	onebot.Service(gatewayChanOnebot)

	gatewayChanClient := gateway.NewGatewayChannel()
	client.Service(gatewayChanClient)

	// Reading channel
	go func() {
		for msg := range gatewayChanClient.ToServer {
			ylog.Debug("Processing [%s]", msg.Content)
			core_copy = *core
			result := core.Process(&msg)

			fmt.Printf("Yachiyo > %s\n", result)

			fmt.Printf("\nDEBUG\nFORMAL:")
			fmt.Println(core_copy.Emotion.String())
			fmt.Println(core_copy.State.Prompt())
			fmt.Printf("\nLATER:\n")
			fmt.Println(core.Emotion.String())
			fmt.Println(core.State.Prompt())

			msg.Content = result

			gatewayChanClient.ToClient <- msg
		}
	}()

	go func() {
		for msg := range gatewayChanOnebot.ToServer {
			ylog.Debug("Processing [%s]", msg.Content)
			core_copy = *core
			result := core.Process(&msg)

			fmt.Printf("Yachiyo > %s\n", result)

			fmt.Printf("\nDEBUG\nFORMAL:")
			fmt.Println(core_copy.Emotion.String())
			fmt.Println(core_copy.State.Prompt())
			fmt.Printf("\nLATER:\n")
			fmt.Println(core.Emotion.String())
			fmt.Println(core.State.Prompt())

			msg.Content = result

			gatewayChanOnebot.ToClient <- msg
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
}
