package gateway

import (
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
)

type Service interface {
	SchemeName() string
	Listen(port int64, submit func(trig trigger.Trigger))
	Send(a action.Action)
}
