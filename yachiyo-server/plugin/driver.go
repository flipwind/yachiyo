package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"yachiyo/yachiyo-utils/logger"

	"go.yaml.in/yaml/v4"
)

type PluginDriver struct {
	mu         sync.Mutex
	pluginPath string
	serverAddr string

	plugins map[string]*PluginRuntime
}

func NewPluginDriver(pluginPath string, serverAddr string) *PluginDriver {
	return &PluginDriver{
		pluginPath: pluginPath,
		serverAddr: serverAddr,
		plugins:    make(map[string]*PluginRuntime),
	}
}

func (d *PluginDriver) Init() {
	// Reading from directions
	// = BuiltinPlugins

	// = ExternalPlugins

	entries, err := os.ReadDir(d.pluginPath)
	if err != nil {
		logger.Error(sourcename, "Plugin base path {%v} reading error: %v", d.pluginPath, err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configPath := filepath.Join(d.pluginPath, entry.Name(), "config.yaml")
		if _, err := os.Stat(configPath); err != nil {
			logger.Error(sourcename, "Cannot read {%v}'s config: %v", configPath, err)
			continue
		}

		configData, err := os.ReadFile(configPath)
		if err != nil {
			logger.Error(sourcename, "Reading {%v}'s config error: %v", configPath, err)
			continue
		}

		var plugincfg PluginConfig
		if err := yaml.Unmarshal(configData, &plugincfg); err != nil {
			logger.Error(sourcename, "{Plugin} Unmarshal Config {%v} error: %v", configPath, err)
			continue
		}

		d.plugins[plugincfg.Name] = &PluginRuntime{
			Name: plugincfg.Name,
			Type: External,
			Directory: filepath.Join(d.pluginPath, entry.Name()),
			Cmd: exec.Command(plugincfg.Runtime, plugincfg.Args...),
			Config: plugincfg,
		}

		logger.Info(sourcename, "Successfully reading {%v}'s config.", plugincfg.Name)
	}

	// Initialize every plugins
	for _, plugin := range d.plugins {
		if plugin.Type == Builtin {

		} else {
			plugin.Cmd.Dir = plugin.Directory

			// TODO: Environment port and address
			plugin.Cmd.Env = append(os.Environ(), "YACHIYO_SERVER_ADDR=" + ":16800")
			
			plugin.Cmd.Stdout = os.Stdout
			plugin.Cmd.Stderr = os.Stderr

			command := plugin.Cmd.String()

			if err := plugin.Cmd.Start(); err != nil {
				logger.Debug(sourcename, "%v", plugin.Directory)
				logger.Error(sourcename, "Config {%v} cannot run {%v}: %v", plugin.Name, command ,err)
				continue
			}
			logger.Success(sourcename, "Config {%v} loaded successfully.", plugin.Name)
		}
	}
}

func (d *PluginDriver) Close() {
	for _, plugin := range d.plugins {
		if plugin.Type != External {
			continue
		}

		logger.Debug(sourcename, "Stopping plugin {%v}", plugin.Name)

		if err := plugin.Cmd.Process.Kill(); err != nil {
			logger.Error(sourcename, "Stopping plugin {%v} failed: %v", plugin.Name, err)
			continue
		}

		if err := plugin.Cmd.Wait(); err != nil {
			logger.Debug(sourcename, "Plugin {%v} killed: %v", plugin.Name, err)
		}
	}
}
