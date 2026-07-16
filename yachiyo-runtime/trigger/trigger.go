package trigger

import "yachiyo/yachiyo-util/logger"

var ylog = logger.New("Yachiyo.Trigger")

type Trigger interface {
	trigger()
}
