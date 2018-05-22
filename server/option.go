package server

import (
	"crypto/tls"
	"time"
)

type option struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	configs      map[string]interface{}
	tlsConfig    *tls.Config
	rpcPath      string
}

func (s *Server) ReadTimeout(timeout time.Duration) *Server {
	s.option.readTimeout = timeout
	return s
}

func (s *Server) WriteTimeout(timeout time.Duration) *Server {
	s.option.writeTimeout = timeout
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
