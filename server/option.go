package server

import (
	"time"
	"crypto/tls"
)

type Option struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	tlsConfig *tls.Config
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
