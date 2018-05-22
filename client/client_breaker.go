package client

import (
	"frpc/circuit"
	"time"
)

type breakerClient struct {
	breaker circuit.Breaker
	timeout time.Duration
}

func (c *Client) Breaker(breaker circuit.Breaker, timeout time.Duration) *Client {
	c.option.breakerClient.breaker = breaker
	c.option.breakerClient.timeout = timeout
	return c
}
