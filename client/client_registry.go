package client

import (
	"frpc/core"
	"frpc/registry"
)

type RegistryClient struct {
	Discovery registry.Discovery
	servers   map[string]string
	ch        chan []*registry.KV
}

func (c *Client) Discovery(mode core.RegistryMode, basePath string, servicePath string, etcdAddr []string) *Client {
	servers := make(map[string]string)
	c.servicePath = servicePath
	innerDiscovery := registry.NewDiscovery(mode, basePath, servicePath, etcdAddr)
	c.cachedClient = make(map[string]*Client)
	c.registry.Discovery = innerDiscovery
	pairs := innerDiscovery.GetServices()
	for _, entry := range pairs {
		servers[entry.Key] = entry.Value
	}
	c.registry.servers = servers
	ch := innerDiscovery.WatchService()
	if ch != nil {
		c.registry.ch = ch
		go c.watch(ch)
	}
	return c
}

func (c *Client) watch(ch chan []*registry.KV) {
	for pairs := range ch {
		servers := make(map[string]string)
		for _, p := range pairs {
			servers[p.Key] = p.Value
		}
		c.mutex.Lock()
		c.registry.servers = servers
		c.mutex.Unlock()
	}
}
