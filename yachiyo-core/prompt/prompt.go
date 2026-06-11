package prompt

import (
	"fmt"
	"time"
	"os"
	"strings"
	"yachiyo/yachiyo-utils/logger"
	"path/filepath"
)

var sourcename string = "Yachiyo.Prompt"

type SystemPrompt struct {
	Content string

	// ResourcePath should be the root path of the resources.
	// Example: "/home/user/package.yachiyo/prompts"
	ResourcePath string

	// Resources should be a list of prompt files' path, except `core.md`.
	// Example: ["a.md", "b.md"]
	Resources []string
}

// Read SystemPrompt resources from the root path of package dir.
func LoadSystemPrompt(path string) SystemPrompt {
	// TODO: Load all the prompt resources.

	info, err := os.Stat(path)
	if err != nil {
		logger.Error(sourcename, "Failed to load system prompt from path %s: %v", path, err)
		return SystemPrompt{}
	}
	if !info.IsDir() {
		logger.Error(sourcename, "Path %s is not a directory", path)
		return SystemPrompt{}
	}

	content, err := os.ReadFile(filepath.Join(path, "core.md"))
	if err != nil {
		logger.Error(sourcename, "Failed to read core.md: %v", err)
		return SystemPrompt{}
	}

	return SystemPrompt{
		Content: string(content),
		ResourcePath: path,
	}
}

func ProcessSystemPrompt(s SystemPrompt, platform string) SystemPrompt {
	// Process the time and platform
	timeStr := time.Now().Format("2006.01.02 15:04:05")
	logger.Debug(sourcename, "Generating the prompt at the time {%s}, using platform {%s}", timeStr, platform)

	s.Content = strings.ReplaceAll(s.Content, "{{time}}", timeStr)
	s.Content = strings.ReplaceAll(s.Content, "{{platform}}", platform)

	logger.Debug(sourcename, "System Prompt Processed: \n%s", s.Content)
	return s
}

func (s SystemPrompt) String() string {
	return fmt.Sprintf("%#v", s)
}