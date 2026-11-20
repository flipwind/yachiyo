package core

import (
	"errors"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-util/yerror"
)

func errorAction(err error, addr address.Address) action.Action {
	code := action.ErrMessageProcess

	switch {
	case errors.Is(err, yerror.ErrLLMChoice):
		code = action.ErrLLMChoiceEmpty
	case errors.Is(err, yerror.ErrLLMGenerate):
		code = action.ErrLLMGenerate
	case errors.Is(err, yerror.ErrMessageStore):
		code = action.ErrMessageStore
	}

	return &action.Error{
		Code:    code,
		Message: err.Error(),
		Address: addr,
	}
}
