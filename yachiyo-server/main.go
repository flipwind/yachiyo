package main

import (
	"time"
	"yachiyo/yachiyo-core/core"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-utils/logger"
)

var log = logger.New("Yachiyo.Server.Main")

func main() {
	log.Info("Initializing Yachiyo server...")

	core := core.New()

	// Simple tests
	result := core.Process(&event.UserMessageEvent{
		Message: event.Message{
			Author: "flipwind",
			Source: "CLI",
			Time: time.Now().Unix(),
			Content: "Hello Yachiyo",
		},
	})

	log.Success("%s", result)
}