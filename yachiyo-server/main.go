package main

import (
	// "bufio"
	// "fmt"
	// "os"
	// "time"
	"fmt"
	"yachiyo/yachiyo-adapter/onebot"
	"yachiyo/yachiyo-core/core"
	"yachiyo/yachiyo-utils/logger"
)

var log = logger.New("Yachiyo.Server.Main")

func main() {
	log.Info("Initializing Yachiyo server...")

	var core_copy core.Core
	core := core.New()

	log.Success("Successfully initialize Yachiyo server.")

	// Loop
	// scanner := bufio.NewScanner(os.Stdin)

	// round := 0

	adapterChanOnebot := onebot.NewAdapterChannel()
	onebot.Service(adapterChanOnebot)

	// Reading channel
	for messageEvent := range adapterChanOnebot.ToServer {
		log.Debug("Processing [%s]", messageEvent.Message.Content)
		result := core.Process(&messageEvent)

		fmt.Printf("Yachiyo > %s\n", result)
		fmt.Printf("\nDEBUG\nFORMAL:")
		// fmt.Println(core.History.ListAll())
		fmt.Println(core_copy.Emotion.String())
		fmt.Println(core_copy.State.Prompt())
		fmt.Printf("\nLATER:\n")
		fmt.Println(core.Emotion.String())
		fmt.Println(core.State.Prompt())

		adapterChanOnebot.ToClient <- onebot.GroupMessageSend{
			GroupID: messageEvent.Message.Payload.(map[string]any)["group_id"].(int64),
			Content: result,
		}
	}

	// for {
	// 	round ++
	// 	fmt.Printf("\n== Round %d ==\n", round)
	// 	core_copy = *core
	// 	fmt.Print("User > ")

	// 	if !scanner.Scan(){
	// 		break
	// 	}

	// 	input := scanner.Text()

	// 	result := core.Process(&event.UserMessageEvent{
	// 		Message: event.Message{
	// 			Type:    "User",
	// 			Author:  "flipwind",
	// 			Source:  "CLI",
	// 			Time:    time.Now().Unix(),
	// 			Content: input,
	// 		},
	// 	})
	// fmt.Printf("Yachiyo > %s\n", result)
	// fmt.Printf("\nDEBUG\nFORMAL:")
	// // fmt.Println(core.History.ListAll())
	// fmt.Println(core_copy.Emotion.String())
	// fmt.Println(core_copy.State.Prompt())
	// fmt.Printf("\nLATER:\n")
	// fmt.Println(core.Emotion.String())
	// fmt.Println(core.State.Prompt())
	// }

	// if err := scanner.Err(); err != nil {
	// 	log.Error("Reading failed: %v", err)
	// }
}
