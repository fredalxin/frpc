package server

import (
	"net"
	"errors"
	kcp "github.com/xtaci/kcp-go"
)

func makeKcpListener(s *Server, address string) (ln net.Listener, err error) {
	if s.configs == nil || s.configs["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in server.Options")
	}
	return kcp.ListenWithOptions(address, s.configs["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

func (s *Server) WithBlockCrypt(bc kcp.BlockCrypt) *Server {
	s.configs["BlockCrypt"] = bc
	return s
}
