package server

import (
	"net"
	"fmt"
)

var makeListeners = make(map[string]MakeListener)

func init() {
	makeListeners["tcp"] = makeTcpListener
	makeListeners["http"] = makeTcpListener
	makeListeners["quic"] = makeQuicListener
	makeListeners["kcp"] = makeKcpListener
}

type MakeListener func(s *Server, address string) (ln net.Listener, err error)


func makeTcpListener(s *Server, address string) (ln net.Listener, err error) {
	return net.Listen("tcp", address)
}

func (s *Server) makeListener(network, address string) (ln net.Listener, err error) {
	ml := makeListeners[network]
	if ml == nil {
		return nil, fmt.Errorf("can not make listener for %s", network)
	}
	return ml(s, address)
}

