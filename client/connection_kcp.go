package client

import (
	"frpc/log"
	kcp "github.com/xtaci/kcp-go"
	"net"
)

func newKcpConn(c *Client, network string, address string) (net.Conn, error) {
	var conn net.Conn
	var err error

	conn, err = kcp.DialWithOptions(address, c.option.block.(kcp.BlockCrypt), 10, 3)

	if err != nil {
		log.Errorf("failed to dial server with kcp: %v", err)
		return nil, err
	}
	return conn, nil
}
