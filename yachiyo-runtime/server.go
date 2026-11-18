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
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

const (
	version = "0.1.0"
)

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

	ylog.Success("Successfully initialize Yachiyo server.")

	if yconfig.Gateways.Onebot.Enabled {
		go serviceChannel(ycore.Pipe, &onebot.OnebotService{}, yconfig.Gateways.Onebot.Port)
	}
	if yconfig.Gateways.Client.Enabled {
		go serviceChannel(ycore.Pipe, &client.ClientService{}, yconfig.Gateways.Client.Port)
	}
	if yconfig.Gateways.JSONClient.Enabled {
		info := jsonclient.ServerInfo{Name: yconfig.Nickname, Version: version}
		go serviceChannel(ycore.Pipe, jsonclient.NewJsonClientService(info), yconfig.Gateways.JSONClient.Port)
	}

	go ycore.Run()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ycore.Close()
	ylog.Info("Shutting down...")
	return nil
}

func serviceChannel(p *core.Pipeline, s gateway.Service, port int64) {
	gatewayChannel := gateway.NewGatewayChannel()
	s.Listen(gatewayChannel, port)

	p.Register(s.SchemeName(), gatewayChannel.ToClient)

	for msg := range gatewayChannel.ToServer {
		switch t := msg.(type) {
		case *trigger.UserMessage:
			ylog.Debug("Processing [%s]", t.Content)
		}
		p.Raw <- msg
	}
}
