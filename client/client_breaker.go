package client

import (
	"frpc/circuit"
	"time"
)

type BreakerClient struct {
	breaker circuit.Breaker
	timeout time.Duration
}

func (c *Client) Breaker(breaker circuit.Breaker, timeout time.Duration) *Client {
	c.option.Breaker.breaker = breaker
	c.option.Breaker.timeout = timeout
	return c
}
