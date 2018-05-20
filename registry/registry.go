package registry

import (
	"frpc/core"
	"time"
)

type Registry interface {
	Start() error
	Register(name string, rcvr interface{}, metadata string) (err error)
}

func NewRegistry(mode core.RegistryMode, basePath string, serviceAddress string, registryAddr []string, interval time.Duration) Registry {
	switch mode {
	case core.Etcd:
		return NewEtcdRegistry(basePath, serviceAddress, registryAddr, interval)
	case core.Consul:
		return NewConsulRegistry(basePath, serviceAddress, registryAddr, interval)
	default:
		return nil
	}
}
