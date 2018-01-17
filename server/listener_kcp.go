package server

import (
	"net"
	"errors"
)

func makeKcpListener(s *Server, address string) (ln net.Listener, err error) {
	return nil, errors.New("kcp unsupported")
}

