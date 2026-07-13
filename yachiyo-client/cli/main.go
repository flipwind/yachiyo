package main

import (
	"fmt"
	"os"
	"yachiyo/yachiyo-client/cli/gateway"
	"yachiyo/yachiyo-client/cli/module"

	tea "charm.land/bubbletea/v2"
)

func main() {
	cliChannel := gateway.NewCliChannel()
	cliGateway := gateway.NewGateway(cliChannel)
	go cliGateway.Listen()

	model := module.InitialModel(cliChannel)
	p := tea.NewProgram(model)

	go func() {
		for msg := range cliChannel.ToClient {
			p.Send(msg)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Running error: %v", err)
		os.Exit(1)
	}
}
