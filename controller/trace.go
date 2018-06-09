package controller

import (
	"context"
	"frpc/protocol"
	"golang.org/x/net/trace"
	"net"
	"net/http"
)

type Trace struct {
}

func NewTrace() *Trace {
	return &Trace{}
}

func (p *Trace) ExportListner(addr string) *Trace {
	go http.ListenAndServe(":8088", nil)
	return p
}

func (p *Trace) Register(name string, rcvr interface{}, metadata string) error {
	tr := trace.New("frpc.Server", "Register")
	defer tr.Finish()
	tr.LazyPrintf("register %s: %T", name, rcvr)
	return nil
}

func (p *Trace) HandleConn(conn net.Conn) (net.Conn, bool) {
	tr := trace.New("frpc.Server", "AcceptConn")
	defer tr.Finish()
	tr.LazyPrintf("accept conn %s", conn.RemoteAddr().String())
	return conn, true
}

func (p *Trace) PostRequest(ctx context.Context, req *protocol.Message, err error) error {
	tr := trace.New("frpc.Server", "ReadRequest")
	defer tr.Finish()
	tr.LazyPrintf("read request %s.%s, seq: %d", req.ServicePath, req.ServiceMethod, req.Seq())
	return nil
}

func (p *Trace) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	tr := trace.New("frpc.Server", "WriteResponse")
	defer tr.Finish()
	if err == nil {
		tr.LazyPrintf("succeed to call %s.%s, seq: %d", req.ServicePath, req.ServiceMethod, req.Seq())
	} else {
		tr.LazyPrintf("failed to call %s.%s, seq: %d : %v", req.Seq, req.ServicePath, req.ServiceMethod, req.Seq(), err)
		tr.SetError()
	}

	return nil
}
