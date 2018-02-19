package registry

import "time"

type Registry interface {
	Start() error
	Register(name string, rcvr interface{}, metadata string) (err error)
}

func NewRegistry(registry string, basePath string, serviceAddress string, etcdAddr []string, interval time.Duration) Registry {
	switch registry {
	case "etcd":
		return NewEtcdRegistry(basePath, serviceAddress, etcdAddr, interval)
	case "consul":
		return nil
	default:
		return nil
	}
}
