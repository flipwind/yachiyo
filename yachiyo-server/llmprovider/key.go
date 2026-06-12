package llmprovider

type Key struct {
	Name string		`yaml:"Name"`
	BaseUrl string	`yaml:"BaseUrl"`
	Secret string	`yaml:"Secret"`
	ModelName string`yaml:"ModelName"`

	StatusEnable bool	`yaml:"StatusEnable"`
}