package main

import (
	"bufio"
	"os"
	"time"
	"yachiyo/yachiyo-core/core"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-utils/logger"
)

var log = logger.New("Yachiyo.Server.Main")

func main() {
	log.Info("Initializing Yachiyo server...")

	core := core.New()

	// Loop
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()

		result := core.Process(&event.UserMessageEvent{
			Message: event.Message{
				Type:    "User",
				Author:  "flipwind",
				Source:  "CLI",
				Time:    time.Now().Unix(),
				Content: input,
			},
		})
		log.Debug("%s", result)

	}

	if err := scanner.Err(); err != nil {
		log.Error("Reading failed: %v", err)
	}
}
