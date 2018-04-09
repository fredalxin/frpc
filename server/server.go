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
	"runtime"
	"net/http"
	"strings"
)

var ErrServerClosed = errors.New("http: Server closed")

const (
	// ReaderBuffsize is used for bufio reader.
	ReaderBuffsize = 1024
	// WriterBuffsize is used for bufio writer.
	WriterBuffsize = 1024
)

type Server struct {
	//待开发 option plugin
	option       Option
	serviceMapMu sync.RWMutex
	serviceMap   map[string]*service
	mu           sync.RWMutex
	doneChan     chan struct{}
	ln           net.Listener
	activeConn   map[net.Conn]struct{}
	registry     RegistryServer
	monitor      MonitorServer

	serviceAddress string
}

func NewServer() *Server {
	return newServer()
}

func newServer() *Server {
	return &Server{
		option: Option{
			configs: make(map[string]interface{}),
		},
	}
}

func (s *Server) ServeProxy() (err error) {
	if s.serviceAddress != "" {
		ss := strings.SplitN(s.serviceAddress, "@", 2)
		if len(ss) == 1 {
			return s.Serve("tcp", s.serviceAddress)
		}
		return s.Serve(ss[0], ss[1])
	}
	errorStr := "server must have serviceAddress first"
	log.Errorf(errorStr)
	return errors.New(errorStr)
}

//func (s *Server) Serve(network, address string) (err error) {
//	return s.ServePath(network, address, "")
//}

func (s *Server) Serve(network, address string) (err error) {
	var ln net.Listener
	ln, err = s.makeListener(network, address)
	if err != nil {
		return
	}
	if network == "http" {
		s.serveHTTPListner(ln, s.option.rpcPath)
		return nil
	}
	return s.serveListener(ln)
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeDoneChanLocked()
	var err error
	if s.ln != nil {
		err = s.ln.Close()
	}
	for c := range s.activeConn {
		c.Close()
		delete(s.activeConn, c)
	}
	return err
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

				log.Errorf("frpc: Accept error: %v; retrying in %v", err, tempDelay)
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

		//todo monitor
		//if s.monitor.metric != (monitor.Metric{}) {
		//	s.monitor.metric.HandleConn(conn)
		//}
		//if s.monitor.trace != (monitor.Trace{}) {
		//	s.monitor.trace.HandleConn(conn)
		//}
		conn, ok := s.monitor.HandleConn(conn)
		if !ok {
			continue
		}

		go s.serveConn(conn)
	}
}

func (s *Server) serveHTTPListner(ln net.Listener, rpcPath string) {
	s.ln = ln

	if rpcPath == "" {
		rpcPath = core.DefaultRPCPath
	}
	http.Handle(rpcPath, s)
	srv := &http.Server{Handler: nil}

	s.mu.Lock()
	if s.activeConn == nil {
		s.activeConn = make(map[net.Conn]struct{})
	}
	s.mu.Unlock()

	srv.Serve(ln)
}

var connected = "200 Connected to frpc"

//http.Handler的实现
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Info("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")

	s.mu.Lock()
	s.activeConn[conn] = struct{}{}
	s.mu.Unlock()
	s.serveConn(conn)
}

func (server *Server) serveConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			//println("err:",err)
			const size = 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			log.Errorf("serving %s panic error: %s, stack:\n %s", conn.RemoteAddr(), err, buf)
		}
		server.mu.Lock()
		delete(server.activeConn, conn)
		server.mu.Unlock()
		conn.Close()
	}()

	r := bufio.NewReaderSize(conn, ReaderBuffsize)
	for {
		t := time.Now()
		//timeout
		if server.option.ReadTimeout != 0 {
			conn.SetReadDeadline(t.Add(server.option.ReadTimeout))
		}
		//decode request
		//可以加一些信息context
		ctx := context.Background()
		req, err := server.decodeRequest(ctx, r)
		if err != nil {
			req = nil
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return
			}
			err = errors.New("rpc: server cannot decode request: " + err.Error())
			return
		}
		//timeout
		if server.option.WriteTimeout != 0 {
			conn.SetWriteDeadline(t.Add(server.option.WriteTimeout))
		}
		//record request time
		ctx = context.WithValue(ctx, core.StartRequestContextKey, time.Now().UnixNano())
		//handle request
		go func() {
			if req.IsHeartbeat() {
				log.Info("server receive heartbeat")
				req.SetMessageType(protocol.Response)
				data := req.Encode()
				conn.Write(data)
				return
			}
			resMetadata := make(map[string]string)
			newCtx := context.WithValue(context.WithValue(ctx, core.ReqMetaDataKey, req.Metadata),
				core.ResMetaDataKey, resMetadata)

			res, err := server.handleRequest(newCtx, req)
			if err != nil {
				log.Warnf("frpc: failed to handle request: %v", err)
			}
			if !req.IsOneway() {
				if len(resMetadata) > 0 { //copy meta in context to request
					meta := res.Metadata
					if meta == nil {
						res.Metadata = resMetadata
					} else {
						for k, v := range resMetadata {
							meta[k] = v
						}
					}
				}
				data := res.Encode()
				conn.Write(data)
			}
			//todo monitor
			//if server.monitor.metric != (monitor.Metric{}) {
			//	server.monitor.metric.PostResponse(ctx, req, res, err)
			//}
			//if server.monitor.trace != (monitor.Trace{}) {
			//	server.monitor.trace.PostResponse(ctx, req, res, err)
			//}
			server.monitor.PostResponse(ctx, req, res, err)
			protocol.FreeMsg(req)
			protocol.FreeMsg(res)
		}()

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

func (s *Server) closeDoneChanLocked() {
	ch := s.getDoneChanLocked()
	select {
	case <-ch:
	default:
		close(ch)
	}
}

func (s *Server) getDoneChanLocked() chan struct{} {
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func (s *Server) decodeRequest(ctx context.Context, r io.Reader) (req *protocol.Message, err error) {
	// pool req?
	req = protocol.GetMsgs()
	err = req.Decode(r)
	//todo monitor
	s.monitor.PostRequest(ctx, req, err)
	return req, err
}

func (s *Server) HandleRequest(ctx context.Context, req *protocol.Message) (resp *protocol.Message, err error) {
	log.Infof("only for test use")
	return s.handleRequest(ctx, req)
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
		err = errors.New("frpc: can't find service " + serviceName)
		return handleError(res, err)
	}

	mtype := service.method[methodName]
	if mtype == nil {
		err = errors.New("frpc: can't find method " + methodName)
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
