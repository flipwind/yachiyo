package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"uuid"

	"go.yaml.in/yaml/v4"
)

var validClientName = regexp.MustCompile(`^[a-zA-Z0-9_]{1,16}$`)

type Config struct {
	Yachiyo YachiyoConfig `yaml:"yachiyo"`
	QQ      QQConfig      `yaml:"qq"`
	Cache   CacheConfig   `yaml:"cache"`
}

type YachiyoConfig struct {
	GatewayURL           string `yaml:"gateway_url"`
	ClientID             string `yaml:"client_id"`
	ClientName           string `yaml:"client_name"`
	HeartbeatIntervalSec int    `yaml:"heartbeat_interval_sec"`
}

type QQConfig struct {
	AppID               string `yaml:"app_id"`
	AppSecret           string `yaml:"app_secret"`
	Sandbox             bool   `yaml:"sandbox"`
	APIBaseURL          string `yaml:"api_base_url,omitempty"`
	DefaultMarkdown     bool   `yaml:"default_markdown"`
	FallbackToPlainText bool   `yaml:"fallback_to_plain_text"`
	GroupReplyMode      string `yaml:"group_reply_mode"`
	LogContent          bool   `yaml:"log_content"`
	QRSource            string `yaml:"qr_source"`
}

type CacheConfig struct {
	DBPath                  string `yaml:"db_path"`
	MaxSizeMB               int    `yaml:"max_size_mb"`
	RecentMessagesPerGroup  int    `yaml:"recent_messages_per_group"`
	DedupTTLSec             int    `yaml:"dedup_ttl_sec"`
	PassiveReplyMaxCount    int    `yaml:"passive_reply_max_count"`
	C2CPassiveReplyMaxCount int    `yaml:"c2c_passive_reply_max_count"`
	PassiveReplyTTLSec      int    `yaml:"passive_reply_ttl_sec"`
}

func Default() *Config {
	return &Config{
		Yachiyo: YachiyoConfig{
			GatewayURL:           "ws://127.0.0.1:16899/ws/",
			ClientID:             "",
			ClientName:           "qq",
			HeartbeatIntervalSec: 15,
		},
		QQ: QQConfig{
			AppID:               "",
			AppSecret:           "",
			Sandbox:             false,
			APIBaseURL:          "",
			DefaultMarkdown:     true,
			FallbackToPlainText: true,
			GroupReplyMode:      "at_only",
			LogContent:          false,
			QRSource:            "",
		},
		Cache: CacheConfig{
			DBPath:                  "qq_cache.db",
			MaxSizeMB:               60,
			RecentMessagesPerGroup:  100,
			DedupTTLSec:             120,
			PassiveReplyMaxCount:    5,
			C2CPassiveReplyMaxCount: 4,
			PassiveReplyTTLSec:      3500,
		},
	}
}

// Load reads config from path (creating defaults if missing), generates a UUID
// when yachiyo.client_id is empty, validates all required fields, and persists
// any auto-generated changes back to path.
func Load(path string) (*Config, error) {
	cfg := Default()
	dirty := false

	raw, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read config %q: %w", path, err)
		}
		dirty = true
	} else {
		if err := yaml.Unmarshal(raw, cfg); err != nil {
			return nil, fmt.Errorf("parse config %q: %w", path, err)
		}
	}

	cfg.applyDefaults()

	cfg.Yachiyo.ClientID = strings.TrimSpace(cfg.Yachiyo.ClientID)
	if cfg.Yachiyo.ClientID == "" {
		cfg.Yachiyo.ClientID = uuid.NewV7().String()
		dirty = true
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if dirty {
		if err := Save(path, cfg); err != nil {
			return nil, fmt.Errorf("save config %q: %w", path, err)
		}
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.Yachiyo.GatewayURL) == "" {
		c.Yachiyo.GatewayURL = "ws://127.0.0.1:16899/ws/"
	}
	if strings.TrimSpace(c.Yachiyo.ClientName) == "" {
		c.Yachiyo.ClientName = "qq"
	}
	if c.Yachiyo.HeartbeatIntervalSec <= 0 || c.Yachiyo.HeartbeatIntervalSec > 20 {
		c.Yachiyo.HeartbeatIntervalSec = 15
	}
	c.QQ.GroupReplyMode = strings.ToLower(strings.TrimSpace(c.QQ.GroupReplyMode))
	if c.QQ.GroupReplyMode == "" {
		c.QQ.GroupReplyMode = "at_only"
	}
	if strings.TrimSpace(c.Cache.DBPath) == "" {
		c.Cache.DBPath = "qq_cache.db"
	}
	if c.Cache.MaxSizeMB <= 0 {
		c.Cache.MaxSizeMB = 60
	}
	if c.Cache.RecentMessagesPerGroup <= 0 {
		c.Cache.RecentMessagesPerGroup = 100
	}
	if c.Cache.DedupTTLSec <= 0 {
		c.Cache.DedupTTLSec = 120
	}
	if c.Cache.PassiveReplyMaxCount <= 0 {
		c.Cache.PassiveReplyMaxCount = 5
	}
	if c.Cache.C2CPassiveReplyMaxCount <= 0 {
		c.Cache.C2CPassiveReplyMaxCount = 4
	}
	if c.Cache.PassiveReplyTTLSec <= 0 {
		c.Cache.PassiveReplyTTLSec = 3500
	}
}

func (c *Config) Validate() error {
	parsed, err := uuid.Parse(c.Yachiyo.ClientID)
	if err != nil {
		return fmt.Errorf("invalid yachiyo.client_id %q: %w", c.Yachiyo.ClientID, err)
	}
	c.Yachiyo.ClientID = parsed.String()

	if !validClientName.MatchString(c.Yachiyo.ClientName) {
		return fmt.Errorf("invalid yachiyo.client_name %q: must match ^[a-zA-Z0-9_]{1,16}$", c.Yachiyo.ClientName)
	}
	if strings.TrimSpace(c.Yachiyo.GatewayURL) == "" {
		return errors.New("yachiyo.gateway_url cannot be empty")
	}
	if c.QQ.GroupReplyMode != "at_only" && c.QQ.GroupReplyMode != "all" {
		return fmt.Errorf("invalid qq.group_reply_mode %q: must be 'at_only' or 'all'", c.QQ.GroupReplyMode)
	}
	return nil
}

func Save(path string, cfg *Config) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}
