package server

import "time"

type Option struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (s *Server) ReadTimeout(timeout time.Duration) *Server {
	s.option.ReadTimeout = timeout
	return s
}

func (s *Server) WriteTimeout(timeout time.Duration) *Server {
	s.option.WriteTimeout = timeout
	return s
}
