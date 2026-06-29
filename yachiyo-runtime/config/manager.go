package config

import (
	"os"
	"go.yaml.in/yaml/v4"
)

type ConfigManager struct {
	CurrentConfig Config
}

func LoadConfig(configPath string) (ConfigManager, error) {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return ConfigManager{}, err
	}

	var config Config
	if err := yaml.Unmarshal(fileData, &config); err != nil {
		return ConfigManager{}, err
	}

	return ConfigManager{
		CurrentConfig: config,
	}, nil
}