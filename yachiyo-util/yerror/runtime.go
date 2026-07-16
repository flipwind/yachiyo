package yerror

import "fmt"

type RuntimeError struct {
	Type   string
	Reason string
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("%v threw an error: %v", e.Type, e.Reason)
}

func TypeMissing(t string) error {
	return &RuntimeError{
		Type:   t,
		Reason: "missing correct object",
	}
}
