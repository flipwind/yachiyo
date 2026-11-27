package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"uuid"

	"yachiyo/yachiyo-client/qq/bridge"
	"yachiyo/yachiyo-client/qq/config"
	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/tools"
	"yachiyo/yachiyo-client/qq/yws"
	"yachiyo/yachiyo-util/logger"
)

var mainLog = logger.New("Yachiyo.QQ")

func main() {
	configPath := flag.String("config", "config.yaml", "path to QQ client YAML config")
	debug := flag.Bool("debug", false, "enable debug logging")
	forceQR := flag.Bool("qr", false, "force Mobile QQ QR code binding to refresh app_id and app_secret")
	flag.Parse()

	if *debug {
		logger.SetLogLevel(logger.Debug)
	}

	// 1. Load config, auto-generate UUID if empty, and strictly validate BEFORE connecting.
	cfg, err := config.Load(*configPath)
	if err != nil {
		mainLog.Error("Config validation failed: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. If app_id or app_secret is missing (or -qr is passed), run embedded QR bind flow.
	prevAppID := strings.TrimSpace(cfg.QQ.AppID)
	if *forceQR || prevAppID == "" || strings.TrimSpace(cfg.QQ.AppSecret) == "" {
		mainLog.Info("QQ bot credentials missing or -qr enabled, starting embedded QR bind flow...")
		creds, err := qq.QRConnect(ctx, qq.QRConnectOptions{
			Source:             cfg.QQ.QRSource,
			DisplayQRToConsole: true,
		})
		if err != nil {
			mainLog.Error("QQ QR bind failed: %v", err)
			os.Exit(1)
		}
		if prevAppID != "" && prevAppID != strings.TrimSpace(creds.AppID) {
			newClientID := uuid.NewV7().String()
			mainLog.Warn("QQ AppID switched (%s -> %s); rotating client_id (%s -> %s) to isolate routing and cache",
				prevAppID, creds.AppID, cfg.Yachiyo.ClientID, newClientID)
			cfg.Yachiyo.ClientID = newClientID
		}
		cfg.QQ.AppID = creds.AppID
		cfg.QQ.AppSecret = creds.AppSecret
		if err := config.Save(*configPath, cfg); err != nil {
			mainLog.Warn("Bound successfully (app_id=%s), but failed to persist config: %v", creds.AppID, err)
		} else {
			mainLog.Success("QQ bot bound and saved to %s (app_id=%s)", *configPath, creds.AppID)
		}
	}

	// 3. Initialize QQ token provider, REST API client, local store, and tools.
	tokens := qq.NewTokenProvider(cfg.QQ.AppID, cfg.QQ.AppSecret, nil)
	tokens.StartAutoRefresh(ctx)

	apiClient := qq.NewAPIClient(tokens, qq.APIClientOptions{
		Sandbox:             cfg.QQ.Sandbox,
		APIBaseURL:          cfg.QQ.APIBaseURL,
		DefaultMarkdown:     cfg.QQ.DefaultMarkdown,
		FallbackToPlainText: cfg.QQ.FallbackToPlainText,
	})

	st, err := store.Open(store.Options{
		AppID:              cfg.QQ.AppID,
		PrevAppID:          prevAppID,
		DBPath:             cfg.Cache.DBPath,
		MaxSizeMB:          cfg.Cache.MaxSizeMB,
		RecentLimit:        cfg.Cache.RecentMessagesPerGroup,
		DedupTTL:           time.Duration(cfg.Cache.DedupTTLSec) * time.Second,
		PassiveMaxCount:    cfg.Cache.PassiveReplyMaxCount,
		C2CPassiveMaxCount: cfg.Cache.C2CPassiveReplyMaxCount,
		PassiveTTL:         time.Duration(cfg.Cache.PassiveReplyTTLSec) * time.Second,
	})
	if err != nil {
		mainLog.Error("Open SQLite store (%s) failed: %v", cfg.Cache.DBPath, err)
		os.Exit(1)
	}
	defer st.Close()

	toolReg := tools.NewRegistry(cfg.Yachiyo.ClientID, apiClient, st)

	var br *bridge.Bridge
	yClient, err := yws.NewClient(yws.Options{
		GatewayURL:        cfg.Yachiyo.GatewayURL,
		ClientID:          cfg.Yachiyo.ClientID,
		ClientName:        cfg.Yachiyo.ClientName,
		HeartbeatInterval: time.Duration(cfg.Yachiyo.HeartbeatIntervalSec) * time.Second,
		Capabilities:      []string{"message.text"},
		Tools:             toolReg.Specs(),
		Handlers: yws.Handlers{
			OnConnected: func(cbCtx context.Context) {
				if br != nil {
					br.FlushPending(cbCtx)
				}
			},
			OnRuntimeMessage: func(cbCtx context.Context, msg yws.RuntimeMessage) {
				if br != nil {
					br.EnqueueRuntimeMessage(cbCtx, msg)
				}
			},
			OnRuntimeError: func(_ context.Context, errMsg yws.RuntimeError) {
				mainLog.Error("Yachiyo runtime error [%s] (path=%s): %s", errMsg.Code, errMsg.Path, errMsg.Message)
			},
			OnToolCall: toolReg.Execute,
		},
	})
	if err != nil {
		mainLog.Error("Initialize wsjson client failed: %v", err)
		os.Exit(1)
	}

	br = bridge.NewWithOptions(apiClient, st, yClient, bridge.Options{
		GroupReplyMode: cfg.QQ.GroupReplyMode,
		LogContent:     cfg.QQ.LogContent,
	})

	qqGateway := qq.NewGatewayClient(apiClient, tokens, qq.EventHandlers{
		OnReady: func(ready qq.ReadyData) {
			br.SetBotIdentity(ready.User.ID, ready.User.Username)
		},
		OnGroupMessage: br.EnqueueGroupMessage,
		OnGroupAddRobot: func(ev qq.GroupRobotEvent) {
			st.SetGroupActive(ev.GroupOpenID, true)
		},
		OnGroupDelRobot: func(ev qq.GroupRobotEvent) {
			st.SetGroupActive(ev.GroupOpenID, false)
		},
		OnGroupMsgReject: func(ev qq.GroupRobotEvent) {
			st.SetGroupRejectProactive(ev.GroupOpenID, true)
		},
		OnGroupMsgReceive: func(ev qq.GroupRobotEvent) {
			st.SetGroupRejectProactive(ev.GroupOpenID, false)
		},
		OnC2CMessage: br.EnqueueC2CMessage,
		OnFriendAdd: func(ev qq.FriendEvent) {
			st.SetFriendActive(ev.OpenID, true)
		},
		OnFriendDel: func(ev qq.FriendEvent) {
			st.SetFriendActive(ev.OpenID, false)
		},
		OnC2CMsgReject: func(ev qq.FriendEvent) {
			st.SetC2CRejectProactive(ev.OpenID, true)
		},
		OnC2CMsgReceive: func(ev qq.FriendEvent) {
			st.SetC2CRejectProactive(ev.OpenID, false)
		},
	})

	mainLog.Info("Starting yachiyo-client/qq (client_id=%s, gateway=%s)", cfg.Yachiyo.ClientID, cfg.Yachiyo.GatewayURL)

	var wg sync.WaitGroup
	fatalErrCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := yClient.Run(ctx); err != nil {
			fatalErrCh <- err
			stop()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		qqGateway.Run(ctx)
	}()

	wg.Wait()

	select {
	case err := <-fatalErrCh:
		mainLog.Error("Fatal gateway error: %v", err)
		os.Exit(1)
	default:
		mainLog.Info("yachiyo-client/qq stopped cleanly")
	}
}
