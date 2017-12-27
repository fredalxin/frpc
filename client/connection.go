package client

import (
	"net"
	"time"
	"bufio"
	"frpc/log"
)

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 16 * 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 16 * 1024
)

func (c *Client) Connect(network, address string) error {
	var conn net.Conn
	var err error
	switch network {
	case "http":
		conn, err = newHttpConn(c, network, address)
	case "kcp":
		conn, err = newKcpConn(c, network, address)
	case "quic":
		conn, err = newQuicConn(c, network, address)
	default: //tcp unix udp
		conn, err = newDirectConn(c, network, address)
	}

	if err == nil && conn != nil {
		if c.option.ReadTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(c.option.ReadTimeout))
		}
		if c.option.WriteTimeout != 0 {
			conn.SetWriteDeadline(time.Now().Add(c.option.WriteTimeout))
		}
		c.conn = conn
		reader := bufio.NewReaderSize(conn, ReaderBuffsize)
		go c.handleResponse(reader)

		if c.option.Heartbeat && c.option.HeartbeatInterval > 0 {
			go c.heartbeat()
		}
	}
	return err
}

func newHttpConn(client *Client, network string, address string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, client.option.ConnTimeout)
	if err != nil {
		log.Errorf("failed to dial server: %v", err)
		return nil, err
	}
	//io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n"
	return conn, nil
}
func newKcpConn(client *Client, network string, address string) (net.Conn, error) {
	panic(client)
}
func newQuicConn(client *Client, network string, address string) (net.Conn, error) {
	panic(client)
}

func newDirectConn(client *Client, network string, address string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, address, client.option.ConnTimeout)
	if err != nil {
		log.Errorf("failed to dial server: %v", err)
		return nil, err
	}
	return conn, nil
}
