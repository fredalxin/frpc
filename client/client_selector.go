package client

import (
	"frpc/selector"
	"frpc/log"
)

func (c *Client) Selector(selectorMode string) *Client {
	servers := c.registry.servers
	if servers == nil {
		log.Errorf("please set registry first")
	}
	c.selector = selector.NewSelector(selectorMode, servers)
	return c
}
