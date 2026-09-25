package address

import (
	"net/url"
	"yachiyo/yachiyo-util/logger"
	"yachiyo/yachiyo-util/yerror"
)

var ylog = logger.New("Yachiyo.Address")

type Address struct {
	Content string
}

func (a *Address) Scheme() (string, error) {
	u, err := url.Parse(a.Content)
	if err != nil {
		ylog.Error("address parsing error: %v", err)
		return "", yerror.TypeMissing("address")
	}
	return u.Scheme, nil
}

func (a *Address) Host() (string, error) {
	u, err := url.Parse(a.Content)
	if err != nil {
		ylog.Error("address parsing error: %v", err)
		return "", yerror.TypeMissing("address")
	}
	return u.Host, nil
}