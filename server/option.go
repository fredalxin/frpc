package server

import (
	"time"
	"crypto/tls"
)

type Option struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	configs  map[string]interface{}
	tlsConfig *tls.Config
	rpcPath string
}

func (s *Server) ReadTimeout(timeout time.Duration) *Server {
	s.option.ReadTimeout = timeout
	return s
}

func (s *Server) WriteTimeout(timeout time.Duration) *Server {
	s.option.WriteTimeout = timeout
	return s
}

func (s *Server) TLSConfig(cfg *tls.Config) *Server {
	s.option.tlsConfig = cfg
	return s
}

func (s *Server) RpcPath(path string) *Server {
	s.option.rpcPath = path
	return s
}
