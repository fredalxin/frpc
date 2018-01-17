package client

import (
	"net"
	"errors"
)

func newKcpConn(client *Client, network string, address string) (net.Conn, error) {
	return nil, errors.New("kcp unsupported")
}
