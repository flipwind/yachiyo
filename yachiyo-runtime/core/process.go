package core

import (
	"errors"
	"time"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-shared/trigger"
	"yachiyo/yachiyo-util/yerror"
)

func (c *Core) Process(e trigger.Trigger) action.Action {
	timeNow := time.Now()

	switch t := e.(type) {
	case *trigger.UserMessage:
		c.mu.Lock()
		c.State.Touch(timeNow)

		c.LLMBusy = true
		c.mu.Unlock()
		a := c.processUserMessage(*t)

		c.mu.Lock()
		c.LLMBusy = false
		c.mu.Unlock()

		return a
	case *trigger.InitiativeMessage:
		c.mu.Lock()
		c.State.Touch(timeNow)

		c.LLMBusy = true
		c.mu.Unlock()
		a := c.processInitiativeMessage(*t)

		c.mu.Lock()
		c.LLMBusy = false
		c.mu.Unlock()

		return a
	case *trigger.RuntimeStateRequest:
		return c.processRuntimeStateRequest(*t)
	case *trigger.MessageHistoryRequest:
		return c.processMessageHistoryRequest(*t)
	default:
		ylog.Error("Process unsupported type: %T", t)
		return nil
	}
}

func processError(err error, addr address.Address) action.Action {
	if errors.Is(err, yerror.ErrLLMChoice) {
		return &action.Error{
			Code:    action.ErrLLMChoiceEmpty,
			Message: err.Error(),
			Address: addr,
		}
	}

	if errors.Is(err, yerror.ErrLLMGenerate) {
		return &action.Error{
			Code:    action.ErrLLMGenerate,
			Message: err.Error(),
			Address: addr,
		}
	}

	if errors.Is(err, yerror.ErrMessageStore) {
		return &action.Error{
			Code:    action.ErrMessageStore,
			Message: err.Error(),
			Address: addr,
		}
	}

	return &action.Error{
		Code:    action.ErrMessageProcess,
		Message: err.Error(),
		Address: addr,
	}
}
