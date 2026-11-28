package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	gateway "yachiyo/yachiyo-gateway"
	"yachiyo/yachiyo-gateway/wsjson"
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/core"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store/sqlite3"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Server.Main")

const (
	version = "0.2.0"
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
		return fmt.Errorf("core loading config failed: %w", err)
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
		return fmt.Errorf("storage created error: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			ylog.Error("Storage close error: %v", err)
		}
	}()

	energy, err := state.NewEnergy(cfg.Initiative.EnergyCurve)
	if err != nil {
		return fmt.Errorf("core loading with invalid curves: %w", err)
	}

	ycore, err := core.New(core.Options{
		Nickname:       cfg.Nickname,
		SystemPrompt:   cfg.Prompt.System,
		CompressPrompt: cfg.Prompt.Compress,
		Personality:    personality,
		Energy:         energy,
	}, core.Deps{
		LLM:   llmClient,
		Store: storage,
		Hub:   hub,
	})
	if err != nil {
		return fmt.Errorf("core loading failed: %w", err)
	}

	if cfg.Gateways.Wsjson.Enabled {
		info := wsjson.ServerInfo{Name: cfg.Nickname, Version: version}
		hub.Attach(wsjson.New(info), cfg.Gateways.Wsjson.Port, ycore.Dispatch)
	}

	ylog.Success("Successfully initialize Yachiyo server.")
	go hub.Run(ctx)
	ycore.Run(ctx)

	ylog.Info("Shutting down...")
	return nil
}
