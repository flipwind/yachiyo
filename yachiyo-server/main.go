package main

import (
	"bufio"
	"fmt"
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

	log.Success("Successfully initialize Yachiyo server.")

	// Loop
	scanner := bufio.NewScanner(os.Stdin)

	round := 0

	for {
		round ++
		fmt.Printf("\n== Round %d ==\n", round)
		fmt.Print("User > ")

		if !scanner.Scan(){
			break
		}

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
		fmt.Printf("Yachiyo > %s\n", result)

		// Status Debug
		fmt.Printf("\nDEBUG\n")
		fmt.Println(core.Emotion.String())
		fmt.Println(core.State.String())

	}

	if err := scanner.Err(); err != nil {
		log.Error("Reading failed: %v", err)
	}
}
