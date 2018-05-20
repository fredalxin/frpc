package server

import (
	"errors"
	kcp "github.com/xtaci/kcp-go"
	"net"
)

func makeKcpListener(s *Server, address string) (ln net.Listener, err error) {
	if s.option.configs == nil || s.option.configs["BlockCrypt"] == nil {
		return nil, errors.New("KCP BlockCrypt must be configured in server.Options")
	}
	return kcp.ListenWithOptions(address, s.option.configs["BlockCrypt"].(kcp.BlockCrypt), 10, 3)
}

func (s *Server) WithBlockCrypt(bc kcp.BlockCrypt) *Server {
	s.option.configs["BlockCrypt"] = bc
	return s
}
