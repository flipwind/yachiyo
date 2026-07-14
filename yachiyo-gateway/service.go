package gateway

type Service interface {
	SchemeName() string
	Listen(c *GatewayChannel, p int64)
}