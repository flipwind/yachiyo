package main

import (
	"fmt"
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

	adapterChanOnebot := onebot.NewAdapterChannel()
	onebot.Service(adapterChanOnebot)

	// Reading channel
	for msg := range adapterChanOnebot.ToServer {
		ylog.Debug("Processing [%s]", msg.Content)
		result := core.Process(&msg)

		fmt.Printf("Yachiyo > %s\n", result)

		fmt.Printf("\nDEBUG\nFORMAL:")
		fmt.Println(core_copy.Emotion.String())
		fmt.Println(core_copy.State.Prompt())
		fmt.Printf("\nLATER:\n")
		fmt.Println(core.Emotion.String())
		fmt.Println(core.State.Prompt())

		msg.Content = result

		adapterChanOnebot.ToClient <- msg
	}
}
