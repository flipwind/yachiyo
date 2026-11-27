package config

import (
	"os"
	"path/filepath"
	"testing"
	"uuid"
)

func TestLoadGeneratesAndPersistsUUIDWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if _, err := uuid.Parse(cfg.Yachiyo.ClientID); err != nil {
		t.Fatalf("expected valid generated UUID, got %q: %v", cfg.Yachiyo.ClientID, err)
	}
	if cfg.Yachiyo.GatewayURL != "ws://127.0.0.1:16899/ws/" {
		t.Fatalf("expected default gateway_url ws://127.0.0.1:16899/ws/, got %q", cfg.Yachiyo.GatewayURL)
	}
	if cfg.QQ.GroupReplyMode != "at_only" {
		t.Fatalf("expected default group_reply_mode=at_only, got %q", cfg.QQ.GroupReplyMode)
	}
	if cfg.Cache.C2CPassiveReplyMaxCount != 4 {
		t.Fatalf("expected C2CPassiveReplyMaxCount=4, got %d", cfg.Cache.C2CPassiveReplyMaxCount)
	}
	if cfg.Cache.RecentMessagesPerGroup != 100 {
		t.Fatalf("expected RecentMessagesPerGroup=100, got %d", cfg.Cache.RecentMessagesPerGroup)
	}

	// Reloading must preserve the exact same UUID.
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("second Load failed: %v", err)
	}
	if cfg2.Yachiyo.ClientID != cfg.Yachiyo.ClientID {
		t.Fatalf("expected persisted UUID %q, got %q", cfg.Yachiyo.ClientID, cfg2.Yachiyo.ClientID)
	}
}

func TestLoadRejectsInvalidUUID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "yachiyo:\n  client_id: \"not-a-valid-uuid\"\n  client_name: \"qq\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected Load to fail on invalid client_id")
	}
}

func TestLoadRejectsInvalidClientName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "yachiyo:\n  client_id: \"\"\n  client_name: \"invalid name!\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected Load to fail on invalid client_name")
	}
}
