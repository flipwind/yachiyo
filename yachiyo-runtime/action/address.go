package action

import "net/url"

type Address struct {
	Content string
}

func (a *Address) Scheme() string {
	u, _ := url.Parse(a.Content)
	return u.Scheme
}