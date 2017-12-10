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
	for  {

		req, err := server.readRequest(context.Background(), r)
		if err != nil {
			req = nil
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return
			}
			err = errors.New("rpc: server cannot decode request: " + err.Error())
			return
		}

	}
}

func (s *Server) readRequest(ctx context.Context, r io.Reader) (req *protocol.Message, err error) {
	// pool req?
	req = protocol.GetMsgs()
	err = req.Decode(r)
	return req, err
}
