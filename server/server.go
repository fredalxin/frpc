package server

import (
	"sync"
	"net"
	"time"
	"errors"
	"frpc/log"
	"bufio"
	"context"
	"io"
	"frpc/protocol"
	"reflect"
	"frpc/core"
	"fmt"
)

var ErrServerClosed = errors.New("http: Server closed")

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 1024
)

type Server struct {
	serviceMapMu sync.RWMutex
	serviceMap   map[string]*service
	mu           sync.RWMutex
	doneChan     chan struct{}
	ln           net.Listener
	activeConn   map[net.Conn]struct{}
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (s *Server) Serve(network, address string) (err error) {
	var ln net.Listener
	ln, err = s.makeListener(network, address)
	if err != nil {
		return
	}
	return s.serveListener(ln)
}

func (s *Server) serveListener(ln net.Listener) error {
	var tempDelay time.Duration
	s.mu.Lock()
	s.ln = ln
	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]struct{})
	}
	s.mu.Unlock()
	for {
		conn, err := ln.Accept()
		if err != nil {
			//deal temp err
			select {
			case <-s.getDoneChan():
				return ErrServerClosed
			default:
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Errorf("rpcx: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.SetKeepAlive(true)
			tc.SetKeepAlivePeriod(3 * time.Minute)
		}

		s.mu.Lock()
		s.activeConn[conn] = struct{}{}
		s.mu.Unlock()

		go s.ServeConn(conn)
	}
}

func (s *Server) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func (server *Server) ServeConn(conn net.Conn) {
	r := bufio.NewReaderSize(conn, ReaderBuffsize)
	for {
		t := time.Now()
		//decode request
		req, err := server.decodeRequest(context.Background(), r)
		if err != nil {
			req = nil
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return
			}
			err = errors.New("rpc: server cannot decode request: " + err.Error())
			return
		}
		//timeout
		if server.writeTimeout != 0 {
			conn.SetWriteDeadline(t.Add(server.writeTimeout))
		}

		//handle request
		go func() {
			res, err := server.handleRequest(context.Background(), req)
			if err != nil {
				log.Warnf("rpcx: failed to handle request: %v", err)
			}
			protocol.FreeMsg(req)
			protocol.FreeMsg(res)
		}()

	}
}

func (s *Server) decodeRequest(ctx context.Context, r io.Reader) (req *protocol.Message, err error) {
	// pool req?
	req = protocol.GetMsgs()
	err = req.Decode(r)
	return req, err
}

func (s *Server) handleRequest(ctx context.Context, req *protocol.Message) (resp *protocol.Message, err error) {
	//pool res
	res := req.Copy()
	res.SetMessageType(protocol.Response)

	serviceName := res.ServicePath
	methodName := res.ServiceMethod

	s.serviceMapMu.RLock()
	service := s.serviceMap[serviceName]
	s.serviceMapMu.RUnlock()

	if service == nil {
		err = errors.New("rpcx: can't find service " + serviceName)
		return handleError(res, err)
	}

	mtype := service.method[methodName]
	if mtype == nil {
		err = errors.New("rpcx: can't find method " + methodName)
		return handleError(res, err)
	}

	var argv, replyv reflect.Value
	// 解析请求中的args
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}

	codec := core.Codecs[req.SerializeType()]
	if codec == nil {
		err = fmt.Errorf("can not find codec for %d", req.SerializeType())
		return handleError(res, err)
	}

	//待开发实现类
	err = codec.Decode(req.Payload, argv.Interface())
	if err != nil {
		return handleError(res, err)
	}

	if argIsValue {
		argv = argv.Elem()
	}

	replyv = reflect.New(mtype.ReplyType.Elem())

	err = service.call(ctx, mtype, argv, replyv)
	if err != nil {
		return handleError(res, err)
	}
	data, err := codec.Encode(replyv.Interface())
	if err != nil {
		return handleError(res, err)
	}
	res.Payload = data

	return res, nil
}

func handleError(res *protocol.Message, err error) (*protocol.Message, error) {
	res.SetMessageStatusType(protocol.Error)
	if res.Metadata == nil {
		res.Metadata = make(map[string]string)
	}
	res.Metadata[protocol.ServiceError] = err.Error()
	return res, err
}
