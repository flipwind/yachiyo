package yerror

import "fmt"

type FieldError struct {
	Field  string
	Reason string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%v threw an error: %v", e.Field, e.Reason)
}

func FieldRequired(field string) error {
	return &FieldError{
		Field:  field,
		Reason: "field is required",
	}
}

func FieldInvalid(field string, reason string) error {
	return &FieldError{
		Field:  field,
		Reason: reason,
	}
}

func FieldIncomplete(field string) error {
	return &FieldError{
		Field: field,
		Reason: "field is incomplete",
	}
}
