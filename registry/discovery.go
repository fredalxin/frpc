package registry

import "frpc/core"

type KV struct {
	Key   string
	Value string
}

type Discovery interface {
	GetServices() []*KV
	WatchService() chan []*KV
	RemoveWatcher(ch chan []*KV)
	Clone(servicePath string) Discovery
	Close()
}

func NewDiscovery(mode core.RegistryMode, basePath string, servicePath string, etcdAddr []string) Discovery {
	switch mode {
	case core.Etcd:
		return NewEtcdDiscovery(basePath, servicePath, etcdAddr)
	case core.Consul:
		return NewConsulDiscovery(basePath, servicePath, etcdAddr)
	default:
		return nil
	}
}
