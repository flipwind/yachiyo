package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/client"
	"yachiyo/yachiyo-gateway/jsonclient"
	"yachiyo/yachiyo-gateway/onebot"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

func main() {
	ylog.Info("Initializing Yachiyo server...")

	ycore, err := core.New()
	if err != nil {
		ylog.Error("Core loading unsuccessfully: %v", err)
		return
	}

	yconfig := ycore.Config

	ylog.Success("Successfully initialize Yachiyo server.")

	if *yconfig.Gateway.Onebot.Enabled {
		go serviceChannel(ycore.Pipe, &onebot.OnebotService{}, *yconfig.Gateway.Onebot.Port)
	}
	if *yconfig.Gateway.Client.Enabled {
		go serviceChannel(ycore.Pipe, &client.ClientService{}, *yconfig.Gateway.Client.Port)
	}
	
	// a experiential feature initially enabled.
	go serviceChannel(ycore.Pipe, jsonclient.NewJsonClientService(), 16899)

	go ycore.Run()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
}

func serviceChannel(p *core.Pipeline, s gateway.Service, port int64) {
	gatewayChannel := gateway.NewGatewayChannel()
	s.Listen(gatewayChannel, port)

	p.Register(s.SchemeName(), gatewayChannel.ToClient)

	for msg := range gatewayChannel.ToServer {
		switch t := msg.(type) {
		case *trigger.Message:
			ylog.Debug("Processing [%s]", t.Content)
		}
		p.Raw <- msg
	}
}
