package client

import (
	"frpc/core"
	"frpc/log"
	"frpc/selector"
)

func (c *Client) Selector(selectorMode core.SelectMode) *Client {
	servers := c.registry.servers
	if servers == nil {
		log.Errorf("please set registry first")
	}
	c.selector = selector.NewSelector(selectorMode, servers)
	return c
}
