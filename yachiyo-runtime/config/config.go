package config

import (
	"fmt"
	"os"
	"slices"
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
	DefaultValue *float64 `yaml:"value"`
	Max          *float64 `yaml:"max"`
	Weight       *float64 `yaml:"weight"`
}

type Config struct {
	Nickname *string `yaml:"nickname"`
	Prompt   struct {
		SystemPromptPath *string `yaml:"system"`
	} `yaml:"prompt"`
	Gateway struct {
		Onebot GatewayConfig `yaml:"onebot"`
		Client GatewayConfig `yaml:"client"`
	} `yaml:"gateway"`
	Initiative struct {
		Threshold *float64 `yaml:"threshold"`
		Factors   struct {
			Sociability FactorConfig `yaml:"sociability"`
			AloneTime   FactorConfig `yaml:"alonetime"`
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
		ylog.Error("Config reading error: %v", err)
		return Config{}, err
	}

	var config Config
	if err := yaml.Unmarshal(fileData, &config); err != nil {
		ylog.Error("Config unmarshal yaml failed: %v", err)
		return Config{}, err
	}

	pass, errs := config.Check()
	for _, err := range errs {
		if ywarning.IsWarning(err) {
			ylog.Warn("%v", err)
		} else {
			ylog.Error("%v", err)
		}
	}

	if !pass {
		return Config{}, yerror.FieldIncomplete("Config")
	}

	for _, prov := range config.LLM.Providers {
		if *config.LLM.DefaultProviderName == *prov.Name {
			config.LLM.DefaultProvider = prov
			break
		}
	}

	return config, nil
}

func (c *Config) Check() (bool, []error) {
	pass := true
	var errs []error

	// nickname
	if c.Nickname == nil {
		errs = append(errs, ywarning.FieldMissing("nickname", "Yachiyo"))
		nickname := "Yachiyo"
		c.Nickname = &nickname
	}

	// prompt
	if c.Prompt.SystemPromptPath == nil {
		pass = false
		errs = append(errs, yerror.FieldRequired("prompt.system"))
	}

	// gateway
	if c.Gateway.Onebot.Enabled == nil || c.Gateway.Onebot.Port == nil {
		pass = false
		errs = append(errs, yerror.FieldIncomplete("gateway.onebot"))
	}

	if c.Gateway.Client.Enabled == nil || c.Gateway.Client.Port == nil {
		pass = false
		errs = append(errs, yerror.FieldIncomplete("gateway.client"))
	}

	if c.Gateway.Onebot.Enabled != nil && c.Gateway.Client.Enabled != nil &&
		*c.Gateway.Onebot.Enabled == false && *c.Gateway.Client.Enabled == false {
		errs = append(errs, ywarning.New("gateway",
			fmt.Sprintf("Gateways are closed. %v may not notice any input.", *c.Nickname)))
	}

	// initiative
	if c.Initiative.Threshold == nil {
		errs = append(errs, ywarning.FieldMissing("initiative.threshold", "0.9"))
		value := 0.9
		c.Initiative.Threshold = &value
	}

	if c.Initiative.Factors.Sociability.DefaultValue == nil || c.Initiative.Factors.Sociability.Max == nil || c.Initiative.Factors.Sociability.Weight == nil {
		pass = false
		errs = append(errs, yerror.FieldIncomplete("initiative.factors.sociability"))
	} else {
		if *c.Initiative.Factors.Sociability.DefaultValue > *c.Initiative.Factors.Sociability.Max {
			pass = false
			errs = append(errs, yerror.FieldInvalid("initiative.factors.sociability", "DefaultValue should less than Max"))
		}
	}

	if c.Initiative.Factors.AloneTime.DefaultValue == nil || c.Initiative.Factors.AloneTime.Max == nil || c.Initiative.Factors.AloneTime.Weight == nil {
		pass = false
		errs = append(errs, yerror.FieldIncomplete("initiative.factors.alonetime"))
	} else {
		if *c.Initiative.Factors.AloneTime.DefaultValue > *c.Initiative.Factors.AloneTime.Max {
			pass = false
			errs = append(errs, yerror.FieldInvalid("initiative.factors.alonetime", "DefaultVaule should less than Max"))
		}
	}

	// llm
	if c.LLM.DefaultProviderName == nil {
		pass = false
		errs = append(errs, yerror.FieldRequired("llm.default"))
	}

	if len(c.LLM.Providers) == 0 {
		pass = false
		errs = append(errs, yerror.FieldRequired("llm.providers"))
	}

	llm_any_enabled := false
	var llm_names []string
	for i, prov := range c.LLM.Providers {
		if prov.Enabled == nil || prov.Name == nil || prov.BaseUrl == nil || prov.Secret == nil || prov.Model == nil {
			pass = false
			errs = append(errs, yerror.FieldIncomplete(
				fmt.Sprintf("llm.providers[%v]", i),
			))
		}
		if prov.Enabled != nil && *prov.Enabled == true {
			llm_any_enabled = true
		}
		if prov.Name != nil {
			llm_names = append(llm_names, *prov.Name)
		}
	}

	if llm_any_enabled == false {
		errs = append(errs, ywarning.New("llm.provider",
			fmt.Sprintf("No providers are enabled. %v may not process any input.", *c.Nickname)))
	}

	if c.LLM.DefaultProviderName != nil {
		if slices.Contains(llm_names, *c.LLM.DefaultProviderName) == false {
			pass = false
			errs = append(errs, yerror.FieldInvalid("llm.default", "should be one of the providers' name"))
		}
	}

	return pass, errs
}
