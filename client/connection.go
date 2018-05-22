package client

import (
	"bufio"
	"errors"
	"frpc/core"
	"frpc/log"
	"io"
	"net"
	"net/http"
	"time"
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
		if c.option.readTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(c.option.readTimeout))
		}
		if c.option.writeTimeout != 0 {
			conn.SetWriteDeadline(time.Now().Add(c.option.writeTimeout))
		}
		c.conn = conn
		c.r = bufio.NewReaderSize(conn, ReaderBuffsize)
		go c.handleResponse()

		if c.option.heartbeat && c.option.heartbeatInterval > 0 {
			go c.heartbeat()
		}
	}
	return err
}

var connected = "200 Connected to frpc"

//http实现
func newHttpConn(client *Client, network string, address string) (net.Conn, error) {
	path := client.option.rpcPath
	if path == "" {
		path = core.DefaultRPCPath
	}
	conn, err := net.DialTimeout("tcp", address, client.option.connTimeout)
	if err != nil {
		log.Errorf("failed to dial server with http: %v", err)
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

//tcp unix等
func newDirectConn(client *Client, network string, address string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, address, client.option.connTimeout)
	if err != nil {
		log.Errorf("failed to dial server with tcp: %v", err)
		return nil, err
	}
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	return conn, nil
}
