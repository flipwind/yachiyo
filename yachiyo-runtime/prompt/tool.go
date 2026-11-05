package prompt

type Tool struct {
	Description string
	Name        string
	Parameters  map[string]ToolParameter //exhaustruct:optional

	Execute func(payload string) (string, error) `json:"-"`
}

type ToolParameter struct {
	Type        string
	Description string
	Required    bool
}

type ToolCall struct {
	ID      string
	Name    string
	Payload string
}
