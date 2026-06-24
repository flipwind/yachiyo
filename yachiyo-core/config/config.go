package config

type LLMKey struct {
	Name         string	`yaml:"Name"`
	BaseUrl      string	`yaml:"BaseUrl"`
	Secret       string	`yaml:"Secret"`
	StatusEnable bool	`yaml:"StatusEnable"`
	ModelName    string `yaml:"ModelName"`
}

type Config struct {
	LLM struct {
		Key []LLMKey `yaml:"key"`
	} `yaml:"llm"`
}
