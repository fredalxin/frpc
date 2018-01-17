package server

import (
	"net"
	"errors"
)

func makeQuicListener(s *Server, address string) (ln net.Listener, err error) {
	return nil, errors.New("quic unsupported")
}

