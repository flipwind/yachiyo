package plugin

import (
	"os/exec"
)

var sourcename = "Yachiyo.PluginConfig"

type PluginConfig struct {
	Name string					`yaml:"name"`
	Author string				`yaml:"author"`
	Version string				`yaml:"version"`
	Description string			`yaml:"description"`

	Runtime string				`yaml:"runtime"`
	Args []string				`yaml:"args"`

	SubscribedEvents []string	`yaml:"subscribed_events"`
	Tools []ToolDefinition		`yaml:"tools"`
}

type ToolDefinition struct {
	Name string							`yaml:"name"`
	Description string					`yaml:"description"`
	Parameters map[string]ToolParameter	`yaml:"parameters"`
}

type ToolParameter struct {
	Type string			`yaml:"type"`
	Description string `yaml:"description"`
}

type PluginRuntime struct {
	Name string
	Type PluginType
	Directory string
	Cmd *exec.Cmd
	Config PluginConfig
}

// Builtin Plugins
type PluginType int
const (
	Builtin PluginType = iota
	External
)