package config

import (
	"yachiyo/yachiyo-server/llmprovider"
)

type Config struct {
	Runtime struct {
		Port int64 `yaml:"Port"`
	} `yaml:"Runtime"`
	LLM struct {
		Key []llmprovider.Key `yaml:"Key"`
	} `yaml:"LLM"`
	Plugin struct {
		Path string `yaml:"Path"`
	} `yaml:"Plugin"`
}