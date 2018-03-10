package server

import (
	"frpc/registry"
	"time"
	"frpc/core"
)

type RegistryServer struct {
	Registry registry.Registry
}

func (s *Server) Registry(mode core.RegistryMode, basePath string, serviceAddress string, registryAddr []string, interval time.Duration) *Server {
	innerRegistry := registry.NewRegistry(mode, basePath, serviceAddress, registryAddr, interval)
	s.registry.Registry = innerRegistry
	innerRegistry.Start()
	return s
}
