package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/client"
	"yachiyo/yachiyo-gateway/onebot"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

func main() {
	ylog.Info("Initializing Yachiyo server...")

	ycore := core.New()
	yconfig := ycore.Config

	ylog.Success("Successfully initialize Yachiyo server.")

	if *yconfig.Gateway.Onebot.Enabled {
		go serviveChannel(ycore.Pipe, &onebot.OnebotService{}, *yconfig.Gateway.Onebot.Port)
	}
	if *yconfig.Gateway.Client.Enabled {
		go serviveChannel(ycore.Pipe, &client.ClientService{}, *yconfig.Gateway.Client.Port)
	}

	go ycore.Run()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
}

func serviveChannel(p *core.Pipeline, s gateway.Service, port int64) {
	gatewayChannel := gateway.NewGatewayChannel()
	s.Listen(gatewayChannel, port)

	p.Register(s.SchemeName(), gatewayChannel.ToClient)

	for msg := range gatewayChannel.ToServer {
		switch t := msg.(type) {
		case *trigger.Message:
			ylog.Debug("Processing [%s]", t.Content)
			p.Raw <- t
		}
	}
}
