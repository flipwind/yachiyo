package main

import (
	"context"
	"fmt"
	"os"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-core/prompt"
	"yachiyo/yachiyo-server/config"
	"yachiyo/yachiyo-server/llmprovider"
	"yachiyo/yachiyo-utils/logger"
)

var sourcename string = "Yachiyo.Main"

func main() {
	// Since this project is under active development, you may find this unavailable.
	// Some assets like prompts are currently private.

	globalMode := os.Getenv("YACHIYO_GLOBAL_MODE")
	charPath := os.Getenv("YACHIYO_CHAR_PATH")

	logger.Warn(sourcename, "Since some assets are private, this project may be unavailable at present.")
	logger.Info(sourcename, "Hello, Yachiyo!")
	logger.Info(sourcename, "Running Yachiyo in %s mode.", globalMode)
	logger.Info(sourcename, "Now character package reading the path {%s}", charPath)

	sysPrompt := prompt.LoadSystemPrompt(charPath)
	sysPrompt = prompt.ProcessSystemPrompt(sysPrompt, "cli")

	configManager := config.NewConfigManager()
	err := configManager.Load("config.yaml")
	if err != nil {
		logger.Error(sourcename, "Reading Config Failed.")
	}

	currentKey := configManager.CurrentConfig.LLM.Key[0]

	depsekProvider := llmprovider.CreateOpenAIProvider(currentKey.BaseUrl, currentKey.Secret, currentKey.ModelName)

	var messages []chat.Message
	messages = append(messages, chat.Message{
		Role: "system",
		Content: sysPrompt.Content,
	})
	messages = append(messages, chat.Message{
		Role: "user",
		Content: "晚上好啊~",
	})
	
	output, err := depsekProvider.ChatStream(context.Background(), messages)
	if err != nil {
		logger.Error(sourcename, "Chatstreaming Failed: %v", err)
		return
	}

	for word := range output {
		fmt.Print(word)
	}
	fmt.Println()
}
