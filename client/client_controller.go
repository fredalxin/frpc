package client

import (
	"frpc/core"
	"frpc/selector"
	"frpc/log"
	"time"
	"frpc/circuit"
)

type controllerClient struct {
	failMode          core.FailMode
	breakerClient     breakerClient
	selector       selector.Selector
}

type breakerClient struct {
	breaker circuit.Breaker
	timeout time.Duration
}

func (c *Client) Breaker(breaker circuit.Breaker, timeout time.Duration) *Client {
	c.controllerClient.breakerClient.breaker = breaker
	c.controllerClient.breakerClient.timeout = timeout
	return c
}

func (c *Client) FailMode(failMode core.FailMode) *Client {
	c.controllerClient.failMode = failMode
	return c
}

func (c *Client) Selector(selectorMode core.SelectMode) *Client {
	servers := c.registryClient.servers
	if servers == nil {
		log.Errorf("please set registryClient first")
	}
	c.controllerClient.selector = selector.NewSelector(selectorMode, servers)
	return c
}
