package server

import (
	"context"
	"frpc/controller"
	err "frpc/error"
	"frpc/protocol"
	"net"
)

type controllerServer struct {
	controllers []controller.Controller
}

func (s *Server) Metric(metric *controller.Metric) *Server {
	s.controllerServer.controllers = append(s.controllerServer.controllers, metric)
	return s
}

func (s *Server) Trace(trace *controller.Trace) *Server {
	s.controllerServer.controllers = append(s.controllerServer.controllers, trace)
	return s
}

func (s *Server) RateLimit(rateLimit *controller.RateLimit) *Server {
	s.controllerServer.controllers = append(s.controllerServer.controllers, rateLimit)
	return s
}

func (m *controllerServer) Register(name string, rcvr interface{}, metadata string) error {
	var es []error
	for _, m := range m.controllers {
		if controller, ok := m.(controller.Controller); ok {
			err := controller.Register(name, rcvr, metadata)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return err.NewMultiError(es)
	}
	return nil
}

func (m *controllerServer) HandleConn(conn net.Conn) (net.Conn, bool) {
	var flag bool
	for _, m := range m.controllers {
		if controller, ok := m.(controller.Controller); ok {
			conn, flag = controller.HandleConn(conn)
			if !flag { //interrupt
				conn.Close()
				return conn, false
			}
		}
	}
	return conn, true
}
func (m *controllerServer) PostRequest(ctx context.Context, req *protocol.Message, err error) error {
	for _, m := range m.controllers {
		if controller, ok := m.(controller.Controller); ok {
			err := controller.PostRequest(ctx, req, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *controllerServer) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	for _, m := range m.controllers {
		if controller, ok := m.(controller.Controller); ok {
			err := controller.PostResponse(ctx, req, res, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
