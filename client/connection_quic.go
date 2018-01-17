package client

import (
	"net"
	"errors"
)

func newQuicConn(client *Client, network string, address string) (net.Conn, error) {
	return nil, errors.New("quic unsupported")
}
