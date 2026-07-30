package address

import (
	"net/url"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Address")

type Address struct {
	Content string
}

func (a *Address) Scheme() string {
	u, err := url.Parse(a.Content)
	if err != nil {
		ylog.Error("address parsing error: %v", err)
	}
	return u.Scheme
}