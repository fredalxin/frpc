package server

import (
	"context"
	"frpc/controller"
	err "frpc/error"
	"frpc/protocol"
	"net"
)

type ControllerServer struct {
	controllers []controller.Controller
}

func (s *Server) Metric(metric *controller.Metric) *Server {
	s.controller.controllers = append(s.controller.controllers, metric)
	return s
}

func (s *Server) Trace(trace *controller.Trace) *Server {
	s.controller.controllers = append(s.controller.controllers, trace)
	return s
}

func (s *Server) RateLimit(rateLimit *controller.RateLimit) *Server {
	s.controller.controllers = append(s.controller.controllers, rateLimit)
	return s
}

func (m *ControllerServer) Register(name string, rcvr interface{}, metadata string) error {
	var es []error
	for _, m := range m.controllers {
		if monitor, ok := m.(controller.Controller); ok {
			err := monitor.Register(name, rcvr, metadata)
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

func (m *ControllerServer) HandleConn(conn net.Conn) (net.Conn, bool) {
	var flag bool
	for _, m := range m.controllers {
		if monitor, ok := m.(controller.Controller); ok {
			conn, flag = monitor.HandleConn(conn)
			if !flag { //interrupt
				conn.Close()
				return conn, false
			}
		}
	}
	return conn, true
}
func (m *ControllerServer) PostRequest(ctx context.Context, req *protocol.Message, err error) error {
	for _, m := range m.controllers {
		if monitor, ok := m.(controller.Controller); ok {
			err := monitor.PostRequest(ctx, req, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *ControllerServer) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	for _, m := range m.controllers {
		if monitor, ok := m.(controller.Controller); ok {
			err := monitor.PostResponse(ctx, req, res, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
