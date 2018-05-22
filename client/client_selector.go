package client

import (
	"frpc/core"
	"frpc/log"
	"frpc/selector"
)

func (c *Client) Selector(selectorMode core.SelectMode) *Client {
	servers := c.registryClient.servers
	if servers == nil {
		log.Errorf("please set registryClient first")
	}
	c.selector = selector.NewSelector(selectorMode, servers)
	return c
}
