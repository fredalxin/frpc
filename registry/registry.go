package registry

import "time"

type Registry interface {
	Start() error
	Register(name string, rcvr interface{}, metadata string) (err error)
}

func NewRegistry(registry string, basePath string, serviceAddress string, registryAddr []string, interval time.Duration) Registry {
	switch registry {
	case "etcd":
		return NewEtcdRegistry(basePath, serviceAddress, registryAddr, interval)
	case "consul":
		return NewConsulRegistry(basePath, serviceAddress, registryAddr, interval)
	default:
		return nil
	}
}
