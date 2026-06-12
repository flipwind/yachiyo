package config

import (
	"yachiyo/yachiyo-server/llmprovider"
)

type Config struct {
	LLM struct {
		Key []llmprovider.Key `yaml:"Key"`
	} `yaml:"LLM"`
}