package client

import (
	"net"
	"time"
	"bufio"
	"frpc/log"
	"io"
	"net/http"
	"errors"
	"frpc/core"
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
		c.r = bufio.NewReaderSize(conn, ReaderBuffsize)
		go c.handleResponse()

		if c.option.Heartbeat && c.option.HeartbeatInterval > 0 {
			go c.heartbeat()
		}
	}
	return err
}

var connected = "200 Connected to frpc"

func newHttpConn(client *Client, network string, address string) (net.Conn, error) {
	path := client.option.RPCPath
	if path == "" {
		path = core.DefaultRPCPath
	}
	conn, err := net.DialTimeout("tcp", address, client.option.ConnTimeout)
	if err != nil {
		log.Errorf("failed to dial server: %v", err)
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return conn, nil
	}
	if err == nil {
		log.Errorf("unexpected HTTP response: %v", err)
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}
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
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	return conn, nil
}
