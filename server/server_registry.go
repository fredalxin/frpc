package server

import (
	"frpc/core"
	"frpc/registry"
	"time"
)

type registryServer struct {
	registry registry.Registry
}

func (s *Server) Registry(mode core.RegistryMode, basePath string, serviceAddress string, registryAddr []string, interval time.Duration) *Server {
	innerRegistry := registry.NewRegistry(mode, basePath, serviceAddress, registryAddr, interval)
	s.serviceAddress = serviceAddress
	s.registryServer.registry = innerRegistry
	innerRegistry.Start()
	return s
}
