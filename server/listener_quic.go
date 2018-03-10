package server

import (
	"net"
)

func makeQuicListener(s *Server, address string) (ln net.Listener, err error) {
	//if s.option.tlsConfig == nil {
	//	return nil, errors.New("TLSConfig must be configured in server.Options")
	//}
	//return quicconn.Listen("udp", address, s.option.tlsConfig)
	return nil,nil
}

