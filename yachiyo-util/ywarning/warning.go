package ywarning

import "errors"

type Warning interface {
	error
	Warning()
}

func IsWarning(err error) bool {
	var w Warning
	return errors.As(err, &w)
}