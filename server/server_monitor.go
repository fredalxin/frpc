package server

import (
	"frpc/monitor"
	"github.com/rcrowley/go-metrics"
	"net"
	"context"
	"frpc/protocol"
	err "frpc/error"
)

type MonitorServer struct {
	//metric   monitor.Metric
	//trace    monitor.Trace
	monitors []monitor.Monitor
}

func (s *Server) Metrics(registry metrics.Registry) *Server {
	//s.monitor.metric = *monitor.NewMetrics(registry)
	s.monitor.monitors = append(s.monitor.monitors, monitor.NewMetrics(registry))
	return s
}

func (s *Server) DefaultMetrics() *Server {
	//s.monitor.metric = *monitor.NewDefaultMetrics()
	s.monitor.monitors = append(s.monitor.monitors, monitor.NewDefaultMetrics())
	return s
}

func (s *Server) Trace(registry metrics.Registry) *Server {
	//s.monitor.trace = *monitor.NewTrace()
	s.monitor.monitors = append(s.monitor.monitors, monitor.NewTrace())
	return s
}

func (m *MonitorServer) Register(name string, rcvr interface{}, metadata string) error {
	var es []error
	for _, m := range m.monitors {
		if monitor, ok := m.(monitor.Monitor); ok {
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

func (m *MonitorServer) HandleConn(conn net.Conn) (net.Conn, bool) {
	var flag bool
	for _, m := range m.monitors {
		if monitor, ok := m.(monitor.Monitor); ok {
			conn, flag = monitor.HandleConn(conn)
			if !flag { //interrupt
				conn.Close()
				return conn, false
			}
		}
	}
	return conn, true
}
func (m *MonitorServer) PostRequest(ctx context.Context, req *protocol.Message, err error) error {
	for _, m := range m.monitors {
		if monitor, ok := m.(monitor.Monitor); ok {
			err := monitor.PostRequest(ctx, req, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *MonitorServer) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	for _, m := range m.monitors {
		if monitor, ok := m.(monitor.Monitor); ok {
			err := monitor.PostResponse(ctx, req, res, err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
