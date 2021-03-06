package controller

import (
	"context"
	"frpc/protocol"
	"net"
)

type Controller interface {
	Register(name string, rcvr interface{}, metadata string) (err error)
	HandleConn(conn net.Conn) (net.Conn, bool)
	PostRequest(ctx context.Context, req *protocol.Message, e error) error
	PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error
}
