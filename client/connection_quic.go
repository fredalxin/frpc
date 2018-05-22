package client

import (
	"crypto/tls"
	"frpc/log"
	quic "github.com/marten-seemann/quic-conn"
	"net"
)

func newQuicConn(client *Client, network string, address string) (net.Conn, error) {
	var conn net.Conn
	var err error

	tlsConf := client.option.tlsConfig
	if tlsConf == nil {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	conn, err = quic.Dial(address, tlsConf)

	if err != nil {
		log.Errorf("failed to dial server with quic: %v", err)
		return nil, err
	}

	return conn, nil
}
