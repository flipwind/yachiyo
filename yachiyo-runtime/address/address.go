package address

import (
	"fmt"
	"net/url"
)

type Address struct {
	Content string
}

func (a *Address) Scheme() (string, error) {
	u, err := url.Parse(a.Content)
	if err != nil {
		return "", fmt.Errorf("address parsing error: %w", err)
	}
	return u.Scheme, nil
}

func (a *Address) Host() (string, error) {
	u, err := url.Parse(a.Content)
	if err != nil {
		return "", fmt.Errorf("address parsing error: %w", err)
	}
	return u.Host, nil
}
