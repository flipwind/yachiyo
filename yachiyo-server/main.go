package main

import (
	"yachiyo/yachiyo-core/core"
	"yachiyo/yachiyo-core/memory"
	"yachiyo/yachiyo-utils/logger"
)

var log = logger.New("Yachiyo.Server.Main")

func main() {
	log.Info("Initializing Yachiyo server...")

	core := core.New()

	// Simple tests
	core.Memory.Remember(memory.Memory{
		Content: "Hello Yachiyo~",
	})

	log.Info("%v", core.Memory.ListAll())
}