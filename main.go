package main

import (
	"os"
	"yachiyo/yachiyo-core/prompt"
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
}
