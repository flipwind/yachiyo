package trigger

import "net/url"

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