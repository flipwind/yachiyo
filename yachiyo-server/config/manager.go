package config

import (
	"os"
	"yachiyo/yachiyo-utils/logger"
	"go.yaml.in/yaml/v4"
)

var sourcename = "ConfigManager"

type ConfigManager struct {
	CurrentConfig *Config
}

func NewConfigManager() *ConfigManager{
	return &ConfigManager{}
}

func (c *ConfigManager) Load(configPath string) error {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error(sourcename, "Reading Config {%v} threw an error: %v", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(fileData, &config); err != nil {
		logger.Error(sourcename, "Unmarshal Config {%v} error: %v", configPath, err)
	}

	c.CurrentConfig = &config
	logger.Success(sourcename, "Reading config {%v} successfully.", configPath)
	logger.Debug(sourcename, "%#v", config)

	return nil
}

func (c *ConfigManager) GetConfig() *Config{
	return c.CurrentConfig
}