package client

import (
	"net"
	"crypto/tls"
	quicconn "github.com/marten-seemann/quic-conn"
)

func newQuicConn(client *Client, network string, address string) (net.Conn, error) {
	var conn net.Conn
	var err error

	tlsConf := client.option.TLSConfig
	if tlsConf == nil {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	conn, err = quicconn.Dial(address, tlsConf)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
