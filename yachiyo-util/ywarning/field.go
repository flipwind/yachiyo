package ywarning

import "fmt"

type FieldWarning struct {
	Field  string
	Reason string
}

func (w FieldWarning) Error() string {
	return fmt.Sprintf("%v threw a warning: %v", w.Field, w.Reason)
}

func (w FieldWarning) Warning() {}

func FieldMissing(field string, defaultValue string) FieldWarning {
	return FieldWarning{
		Field: field,
		Reason: fmt.Sprintf("undefined, default value to [%v]", defaultValue),
	}
}

func New(field string, reason string) FieldWarning {
	return FieldWarning{
		Field: field,
		Reason: reason,
	}
}
