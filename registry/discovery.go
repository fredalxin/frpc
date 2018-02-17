package registry

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

func NewDiscovery(discovery string, basePath string, servicePath string, etcdAddr []string) Discovery {
	switch discovery {
	case "etcd":
		return NewEtcdDiscovery(basePath, servicePath, etcdAddr)
	case "consul":
		return NewConsulDiscovery(basePath, servicePath, etcdAddr)
	default:
		return nil
	}
}
