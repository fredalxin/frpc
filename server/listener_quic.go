package server

import (
	"net"
	quic "github.com/marten-seemann/quic-conn"
	"errors"
	"crypto/tls"
)

func makeQuicListener(s *Server, address string) (ln net.Listener, err error) {
	if s.option.tlsConfig == nil {
		return nil, errors.New("TLSConfig must be configured in server.Options")
	}
	return quic.Listen("udp", address, s.option.tlsConfig)
}

func (s *Server) WithTlsConfig(tls *tls.Config) *Server {
	s.option.tlsConfig = tls
	return s
}
