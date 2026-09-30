package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/client"
	"yachiyo/yachiyo-gateway/jsonclient"
	"yachiyo/yachiyo-gateway/onebot"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-runtime/ycontext"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

func main() {
	if err := run(); err != nil {
		ylog.Error("Yachiyo server exit with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ylog.Info("Initializing Yachiyo server...")

	ycore, err := core.New()
	if err != nil {
		return fmt.Errorf("Core loading failed: %w", err)
	}

	yconfig := ycore.Config

	mutableContext := ycontext.Generate(func(c *ycontext.Context){
		c.Name = *ycore.Config.Nickname
	})

	ylog.Success("Successfully initialize Yachiyo server.")

	if *yconfig.Gateway.Onebot.Enabled {
		go serviceChannel(ycore.Pipe, &onebot.OnebotService{}, *yconfig.Gateway.Onebot.Port)
	}
	if *yconfig.Gateway.Client.Enabled {
		go serviceChannel(ycore.Pipe, &client.ClientService{}, *yconfig.Gateway.Client.Port)
	}
	if *yconfig.Gateway.JsonClient.Enabled {
		go serviceChannel(ycore.Pipe, jsonclient.NewJsonClientService(mutableContext.GetContext), *yconfig.Gateway.JsonClient.Port)
	}

	go ycore.Run()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
	return nil
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
