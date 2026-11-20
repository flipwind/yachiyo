package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	gateway "yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/client"
	"yachiyo/yachiyo-gateway/jsonclient"
	"yachiyo/yachiyo-gateway/onebot"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store/sqlite3"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

const (
	version = "0.1.0"
)

func main() {
	if err := run(); err != nil {
		ylog.Error("Yachiyo server exit with error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ylog.Info("Initializing Yachiyo server...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub := gateway.NewHub()

	// TODO: change config status (dev)
	cfg, warns, err := config.Load("config.yaml")
	for _, wa := range warns {
		ylog.Warn("Config warning: %v", wa)
	}
	if err != nil {
		return fmt.Errorf("Core loading config failed: %w", err)
	}

	if cfg.LogLevel == config.LogDebug {
		logger.SetLogLevel(logger.Debug)
	} else {
		logger.SetLogLevel(logger.Info)
	}

	// TODO: LLM List
	llmClient := provider.NewOpenAIProvider(cfg.LLMProvider.BaseURL, cfg.LLMProvider.Secret, cfg.LLMProvider.Model)

	personality := state.BigFive{
		Openness:          cfg.BigFive.Openness,
		Conscientiousness: cfg.BigFive.Conscientiousness,
		Extraversion:      cfg.BigFive.Extraversion,
		Agreeableness:     cfg.BigFive.Agreeableness,
		Neuroticism:       cfg.BigFive.Neuroticism,
	}

	storage, err := sqlite3.Open("yachiyo.db")
	if err != nil {
		return fmt.Errorf("Storage created error: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			ylog.Error("Storage close error: %v", err)
		}
	}()

	energy, err := state.NewEnergy(cfg.Initiative.EnergyCurve)
	if err != nil {
		return fmt.Errorf("Core loading with invalid curves: %w", err)
	}

	ycore, err := core.New(core.Options{
		SystemPrompt:   cfg.Prompt.System,
		CompressPrompt: cfg.Prompt.Compress,
		Personality:    personality,
		Energy:         energy,
	}, core.Deps{
		LLM:     llmClient,
		Store:   storage,
		Deliver: hub.Deliver,
	})
	if err != nil {
		return fmt.Errorf("Core loading failed: %w", err)
	}

	if cfg.Gateways.Onebot.Enabled {
		hub.Attach(onebot.New(), cfg.Gateways.Onebot.Port, ycore.Dispatch)
	}
	if cfg.Gateways.Client.Enabled {
		hub.Attach(client.New(), cfg.Gateways.Client.Port, ycore.Dispatch)
	}
	if cfg.Gateways.JSONClient.Enabled {
		info := jsonclient.ServerInfo{Name: cfg.Nickname, Version: version}
		hub.Attach(jsonclient.New(info), cfg.Gateways.JSONClient.Port, ycore.Dispatch)
	}

	ylog.Success("Successfully initialize Yachiyo server.")
	go hub.Run(ctx)
	ycore.Run(ctx)

	ylog.Info("Shutting down...")
	return nil
}
