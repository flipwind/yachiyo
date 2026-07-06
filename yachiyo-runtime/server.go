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
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

func main() {
	ylog.Info("Initializing Yachiyo server...")

	core := core.New()
	pipeline := core.NewPipeline()

	ylog.Success("Successfully initialize Yachiyo server.")

	go pipeline.Listen()
	go pipeline.DistributionListen()
	go serviveChannel(pipeline, &onebot.OnebotService{})
	go serviveChannel(pipeline, &client.ClientService{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	ylog.Info("Shutting down...")
}

func serviveChannel(p *core.Pipeline, s gateway.Service) {
	gatewayChannel := gateway.NewGatewayChannel()
	s.Listen(gatewayChannel)

	p.Register(s.SchemeName(), &gatewayChannel.ToClient)

	for msg := range gatewayChannel.ToServer {
		ylog.Debug("Processing [%s]", msg.Content)
		p.Raw <- &msg
	}
}
