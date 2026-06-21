package fake

type FakeLLM struct{}

func NewFakeLLM() *FakeLLM {
	return &FakeLLM{}
}

func (l *FakeLLM) LLM() {}

func (l *FakeLLM) Gen(content string) string {
	return "Hello from Yachiyo: \n" + content
}