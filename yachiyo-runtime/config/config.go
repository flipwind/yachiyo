package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/yerror"
	"yachiyo/yachiyo-util/ywarning"

	"go.yaml.in/yaml/v4"
)

var ylog = logger.New("Yachiyo.Config")

type LLMProvider struct {
	Enabled *bool   `yaml:"enabled"`
	Name    *string `yaml:"name"`
	BaseUrl *string `yaml:"base_url"`
	Secret  *string `yaml:"secret"`
	Model   *string `yaml:"model"`
}

type GatewayConfig struct {
	Enabled *bool  `yaml:"enabled"`
	Port    *int64 `yaml:"port"`
}

type FactorConfig struct {
	Curve        *string  `yaml:"curve"`
	DefaultValue *float64 `yaml:"value"`
	Max          *float64 `yaml:"max"`
	Weight       *float64 `yaml:"weight"`
}

type Config struct {
	Nickname *string `yaml:"nickname"`
	Log      struct {
		Level *string `yaml:"level"`
	} `yaml:"log"`
	Prompt struct {
		SystemPromptPath *string `yaml:"system"`
		SystemPrompt     string
	} `yaml:"prompt"`
	Gateway struct {
		Onebot     GatewayConfig `yaml:"onebot"`
		Client     GatewayConfig `yaml:"client"`
		JsonClient GatewayConfig `yaml:"jsonclient"`
	} `yaml:"gateway"`
	Initiative struct {
		Threshold *float64 `yaml:"threshold"`
		Factors   struct {
			Sociability FactorConfig `yaml:"sociability"`
			AloneTime   FactorConfig `yaml:"alonetime"`
			Daytime     FactorConfig `yaml:"daytime"`
		} `yaml:"factors"`
	}
	LLM struct {
		DefaultProvider     LLMProvider
		DefaultProviderName *string       `yaml:"default"`
		Providers           []LLMProvider `yaml:"providers"`
	} `yaml:"llm"`
}

func LoadConfig(configPath string) (Config, error) {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(fileData, &config); err != nil {
		return Config{}, fmt.Errorf("Config unmarshal yaml failed: %w", err)
	}

	warns, errs := config.Check()
	for _, warn := range warns {
		ylog.Warn("%v", warn)
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("Config invalid: %w", err)
	}

	for _, prov := range config.LLM.Providers {
		if *config.LLM.DefaultProviderName == *prov.Name {
			config.LLM.DefaultProvider = prov
			break
		}
	}

	systemPrompt, err := os.ReadFile(*config.Prompt.SystemPromptPath)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to read system prompt: %w", err)
	}

	config.Prompt.SystemPrompt = string(systemPrompt)

	return config, nil
}

func (c *Config) Check() ([]ywarning.Warning, []error) {
	warns := make([]ywarning.Warning, 0)
	errs := make([]error, 0)

	// nickname
	if c.Nickname == nil {
		warns = append(warns, ywarning.FieldMissing("nickname", "Yachiyo"))
		nickname := "Yachiyo"
		c.Nickname = &nickname
	}

	// loglevel
	if c.Log.Level == nil {
		warns = append(warns, ywarning.FieldMissing("log.level", "info"))
		logger.SetLogLevel(logger.Info)
	} else {
		level := strings.ToLower(*c.Log.Level)

		switch level {
		case "info":
			logger.SetLogLevel(logger.Info)
		case "debug":
			logger.SetLogLevel(logger.Debug)
		default:
			warns = append(warns, ywarning.FieldIncorrect("log.level", "info"))
			logger.SetLogLevel(logger.Info)
		}
	}

	// prompt
	if c.Prompt.SystemPromptPath == nil {
		errs = append(errs, fmt.Errorf("prompt.system: %w", yerror.ErrFieldRequired))
	}

	// gateway
	if c.Gateway.Onebot.Enabled == nil || c.Gateway.Onebot.Port == nil {
		errs = append(errs, fmt.Errorf("gateway.onebot: %w", yerror.ErrFieldIncomplete))
	}

	if c.Gateway.Client.Enabled == nil || c.Gateway.Client.Port == nil {
		errs = append(errs, fmt.Errorf("gateway.client: %w", yerror.ErrFieldIncomplete))
	}

	if c.Gateway.JsonClient.Enabled == nil || c.Gateway.JsonClient.Port == nil {
		errs = append(errs, fmt.Errorf("gateway.jsonclient: %w", yerror.ErrFieldIncomplete))
	}

	if c.Gateway.Onebot.Enabled != nil && c.Gateway.Client.Enabled != nil && c.Gateway.JsonClient.Enabled != nil &&
		*c.Gateway.Onebot.Enabled == false && *c.Gateway.Client.Enabled == false && *c.Gateway.JsonClient.Enabled == false {
		warns = append(warns, ywarning.New("gateway",
			fmt.Sprintf("Gateways are closed. %v may not notice any input.", *c.Nickname)))
	}

	// initiative
	if c.Initiative.Threshold == nil {
		warns = append(warns, ywarning.FieldMissing("initiative.threshold", "0.9"))
		value := 0.9
		c.Initiative.Threshold = &value
	}

	factorConfigs := map[string]*FactorConfig{
		"sociability": &c.Initiative.Factors.Sociability,
		"alonetime":   &c.Initiative.Factors.AloneTime,
		"daytime":     &c.Initiative.Factors.Daytime,
	}

	for name, factorConfig := range factorConfigs {
		err := factorConfig.Check(name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// TODO: factor curve verify

	// llm
	if c.LLM.DefaultProviderName == nil {
		errs = append(errs, fmt.Errorf("llm.default: %w", yerror.ErrFieldRequired))
	}

	if len(c.LLM.Providers) == 0 {
		errs = append(errs, fmt.Errorf("llm.providers: %w", yerror.ErrFieldIncomplete))
	}

	llm_any_enabled := false
	var llm_names []string
	for i, prov := range c.LLM.Providers {
		if prov.Enabled == nil || prov.Name == nil || prov.BaseUrl == nil || prov.Secret == nil || prov.Model == nil {
			errs = append(errs, fmt.Errorf("llm.providers[%v]: %w", i, yerror.ErrFieldIncomplete))
		}
		if prov.Enabled != nil && *prov.Enabled == true {
			llm_any_enabled = true
		}
		if prov.Name != nil {
			llm_names = append(llm_names, *prov.Name)
		}
	}

	if llm_any_enabled == false {
		warns = append(warns, ywarning.New("llm.provider",
			fmt.Sprintf("No providers are enabled. %v may not process any input.", *c.Nickname)))
	}

	if c.LLM.DefaultProviderName != nil {
		if slices.Contains(llm_names, *c.LLM.DefaultProviderName) == false {
			errs = append(errs, fmt.Errorf("llm.default should be one of the providers' name: %w", yerror.ErrFieldInvalid))
		}
	}

	return warns, errs
}

func (f *FactorConfig) Check(name string) error {
	if f.Curve == nil || f.DefaultValue == nil || f.Max == nil || f.Weight == nil {
		return fmt.Errorf("initiative.factors.%s: %w", name, yerror.ErrFieldIncomplete)
	} else {
		if *f.DefaultValue > *f.Max {
			return fmt.Errorf("initiative.factors.%s: %w, DefaultValue should be less than Max", name, yerror.ErrFieldInvalid)
		}
	}

	return nil
}
