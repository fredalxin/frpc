package server

import (
	"frpc/registry"
	"time"
)

type RegistryServer struct {
	Registry registry.Registry
}

func (s *Server) Registry(registryStr string, basePath string, serviceAddress string, etcdAddr []string, interval time.Duration) *Server {
	innerRegistry := registry.NewRegistry(registryStr, basePath, serviceAddress, etcdAddr, interval)
	s.registry.Registry = innerRegistry
	innerRegistry.Start()
	//to do
	return nil
}
