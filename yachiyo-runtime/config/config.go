package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"yachiyo/yachiyo-util/yerror"
	"yachiyo/yachiyo-util/ywarning"

	"go.yaml.in/yaml/v4"
)

type fileBigfive struct {
	Openness          *float64 `yaml:"openness"`
	Conscientiousness *float64 `yaml:"conscientiousness"`
	Extraversion      *float64 `yaml:"extraversion"`
	Agreeableness     *float64 `yaml:"agreeableness"`
	Neuroticism       *float64 `yaml:"neuroticism"`
}

type fileGateway struct {
	Enabled *bool  `yaml:"enabled"`
	Port    *int64 `yaml:"port"`
}

type fileFactor struct {
	Curve  *string  `yaml:"curve"`
	Value  *float64 `yaml:"value"`
	Max    *float64 `yaml:"max"`
	Weight *float64 `yaml:"weight"`
}

type fileProvider struct {
	Enabled *bool   `yaml:"enabled"`
	Name    *string `yaml:"name"`
	BaseUrl *string `yaml:"base_url"`
	Secret  *string `yaml:"secret"`
	Model   *string `yaml:"model"`
}

type file struct {
	Nickname *string `yaml:"nickname"`
	Log      struct {
		Level *string `yaml:"level"`
	} `yaml:"log"`
	Prompt struct {
		System   *string `yaml:"system"`
		Compress *string `yaml:"compress"`
	} `yaml:"prompt"`
	Core struct {
		Bigfive fileBigfive `yaml:"bigfive"`
	} `yaml:"core"`
	Gateway struct {
		JSONClient fileGateway `yaml:"jsonclient"`
	} `yaml:"gateway"`
	Initiative struct {
		Energy struct {
			Curve *string `yaml:"curve"`
		} `yaml:"energy"`
		Threshold *float64 `yaml:"threshold"`
		Factors   struct {
			Sociability fileFactor `yaml:"sociability"`
			Alonetime   fileFactor `yaml:"alonetime"`
			Daytime     fileFactor `yaml:"daytime"`
		} `yaml:"factors"`
	} `yaml:"initiative"`
	LLM struct {
		DefaultProvider *string        `yaml:"default"`
		Providers       []fileProvider `yaml:"providers"`
	} `yaml:"llm"`
}

type LogLevel string

const (
	LogInfo  LogLevel = "info"
	LogDebug LogLevel = "debug"
)

type BigFive struct {
	Openness, Conscientiousness, Extraversion, Agreeableness, Neuroticism float64
}

type Prompt struct {
	System, Compress string
}

type Gateway struct {
	Enabled bool
	Port    int64
}

type Gateways struct {
	JSONClient Gateway
}

type Initiative struct {
	Threshold   float64
	EnergyCurve string
}

type LLMProvider struct {
	Enabled bool
	Name    string
	BaseURL string
	Secret  string
	Model   string
}

type Config struct {
	Nickname    string
	LogLevel    LogLevel
	BigFive     BigFive
	Prompt      Prompt
	Gateways    Gateways
	Initiative  Initiative
	LLMProvider LLMProvider
}

var defaultBigFive = BigFive{
	Openness:          0.92,
	Conscientiousness: 0.65,
	Extraversion:      0.88,
	Agreeableness:     0.82,
	Neuroticism:       0.42,
}

const (
	defaultName      = "Yachiyo"
	defaultThreshold = 0.9
)

func Load(path string) (Config, []ywarning.Warning, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, nil, fmt.Errorf("Failed to read config: %w", err)
	}

	config, warns, err := Parse(data)
	if err != nil {
		return Config{}, warns, err
	}

	system, err := os.ReadFile(config.Prompt.System)
	if err != nil {
		return Config{}, warns, fmt.Errorf("Failed to read system prompt: %w", err)
	}
	config.Prompt.System = string(system)

	compress, err := os.ReadFile(config.Prompt.Compress)
	if err != nil {
		return Config{}, warns, fmt.Errorf("Failed to read compress prompt: %w", err)
	}
	config.Prompt.Compress = string(compress)

	return config, warns, nil
}

func Parse(data []byte) (Config, []ywarning.Warning, error) {
	var f file
	var config Config
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Config{}, nil, fmt.Errorf("Config unmarshal error: %w", err)
	}

	warns := make([]ywarning.Warning, 0)
	errs := make([]error, 0)

	required := func(field string) {
		errs = append(errs, fmt.Errorf("%v: %w", field, yerror.ErrFieldRequired))
	}

	incomplete := func(field string) {
		errs = append(errs, fmt.Errorf("%v: %w", field, yerror.ErrFieldIncomplete))
	}

	missing := func(field string, defaultValue string) {
		warns = append(warns, ywarning.FieldMissing(field, defaultValue))
	}

	// nickname
	config.Nickname = defaultName
	if f.Nickname == nil {
		missing("nickname", defaultName)
	} else {
		config.Nickname = *f.Nickname
	}

	// log.level
	config.LogLevel = LogInfo
	if f.Log.Level == nil {
		missing("log.level", string(LogInfo))
	} else {
		switch LogLevel(strings.ToLower(*f.Log.Level)) {
		case LogInfo:
		case LogDebug:
			config.LogLevel = LogDebug
		default:
			warns = append(warns, ywarning.FieldIncorrect("log.level", string(LogInfo)))
		}
	}

	// prompt
	if f.Prompt.System == nil {
		required("prompt.system")
	} else {
		config.Prompt.System = *f.Prompt.System
	}

	if f.Prompt.Compress == nil {
		required("prompt.compress")
	} else {
		config.Prompt.Compress = *f.Prompt.Compress
	}

	// core.bigfive
	config.BigFive = defaultBigFive
	fields := []*float64{
		f.Core.Bigfive.Openness,
		f.Core.Bigfive.Conscientiousness,
		f.Core.Bigfive.Extraversion,
		f.Core.Bigfive.Agreeableness,
		f.Core.Bigfive.Neuroticism,
	}
	missingFields := 0
	for _, field := range fields {
		if field == nil {
			missingFields++
		}
	}

	if missingFields == 5 {
		missing("core.bigfive", fmt.Sprintf("%#v", defaultBigFive))
	} else if missingFields > 0 && missingFields < 5 {
		incomplete("core.bigfive")
	} else {
		config.BigFive = BigFive{
			Openness:          *f.Core.Bigfive.Openness,
			Conscientiousness: *f.Core.Bigfive.Conscientiousness,
			Extraversion:      *f.Core.Bigfive.Extraversion,
			Agreeableness:     *f.Core.Bigfive.Agreeableness,
			Neuroticism:       *f.Core.Bigfive.Neuroticism,
		}

		parts := []struct {
			name  string
			value float64
		}{
			{"openness", config.BigFive.Openness},
			{"conscientiousness", config.BigFive.Conscientiousness},
			{"extraversion", config.BigFive.Extraversion},
			{"agreeableness", config.BigFive.Agreeableness},
			{"neuroticism", config.BigFive.Neuroticism},
		}

		for _, p := range parts {
			if p.value < 0 || p.value > 1 {
				errs = append(errs, fmt.Errorf("core.bigfive.%s should be in [0, 1], found %v: %w", p.name, p.value, yerror.ErrFieldInvalid))
			}
		}
	}

	// gateway
	gateway := func(name string, g fileGateway) Gateway {
		if g.Enabled == nil || g.Port == nil {
			incomplete(name)
			return Gateway{Enabled: false, Port: 0}
		}
		return Gateway{Enabled: *g.Enabled, Port: *g.Port}
	}

	config.Gateways = Gateways{
		JSONClient: gateway("gateway.jsonclient", f.Gateway.JSONClient),
	}

	gatewayAnyEnabled := config.Gateways.JSONClient.Enabled
	if !gatewayAnyEnabled {
		warns = append(warns, ywarning.New("gateway",
			fmt.Sprintf("Gateways are closed. %v may not notice any input.", config.Nickname)))
	}

	// initiative.energy
	if f.Initiative.Energy.Curve == nil {
		required("initiative.energy")
	} else {
		config.Initiative.EnergyCurve = *f.Initiative.Energy.Curve
	}

	// initiative.threshold
	config.Initiative.Threshold = defaultThreshold
	if f.Initiative.Threshold == nil {
		missing("initiative.threshold", fmt.Sprint(defaultThreshold))
	} else {
		config.Initiative.Threshold = *f.Initiative.Threshold
	}

	// llm
	if f.LLM.DefaultProvider == nil {
		required("llm.default")
	}

	if len(f.LLM.Providers) == 0 {
		required("llm.providers")
	}

	defaultProviderFound := false

	for i, p := range f.LLM.Providers {
		if p.Enabled == nil || p.Name == nil || p.BaseUrl == nil || p.Secret == nil || p.Model == nil {
			incomplete(fmt.Sprintf("llm.providers[%v]", i))
			continue
		}
		if *p.Enabled && f.LLM.DefaultProvider != nil &&
			*f.LLM.DefaultProvider == *p.Name {
			config.LLMProvider = LLMProvider{
				Enabled: true,
				Name:    *p.Name,
				BaseURL: *p.BaseUrl,
				Secret:  *p.Secret,
				Model:   *p.Model,
			}
			defaultProviderFound = true
			break
		}
	}

	if !defaultProviderFound {
		errs = append(errs, fmt.Errorf("llm.providers: No available default provider is provided. %v may not process any input.", config.Nickname))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, warns, fmt.Errorf("Config invalid: %w", err)
	}

	return config, warns, nil
}
