package monitor

import (
	"github.com/rcrowley/go-metrics"
	"github.com/vrischmann/go-metrics-influxdb"
	"time"
	"net"
	"context"
	"frpc/protocol"
	"frpc/server"
)

type Metric struct {
	Registry metrics.Registry
	Prefix   string
}

func NewMetrics(registry metrics.Registry) *Metric {
	return &Metric{Registry: registry}
}

func NewDefaultMetrics() *Metric {
	return &Metric{Registry: metrics.DefaultRegistry}
}

func (m *Metric) withPrefix(s string) string {
	return m.Prefix + s
}

func (m *Metric) Register(name string, rcvr interface{}, metadata string) error {
	serviceCounter := metrics.GetOrRegisterCounter(m.withPrefix("serviceCounter"), m.Registry)
	serviceCounter.Inc(1)
	return nil
}

// HandleConnAccept handles connections from clients
func (p *Metric) HandleConn(conn net.Conn) (net.Conn, bool) {
	clientMeter := metrics.GetOrRegisterMeter(p.withPrefix("clientMeter"), p.Registry)
	clientMeter.Mark(1)
	return conn, true
}

// PostReadRequest counts read
func (p *Metric) PostRequest(ctx context.Context, req *protocol.Message, e error) error {
	sp := req.ServicePath
	sm := req.ServiceMethod

	if sp == "" {
		return nil
	}
	m := metrics.GetOrRegisterMeter(p.withPrefix("service."+sp+"."+sm+".Read_Qps"), p.Registry)
	m.Mark(1)
	return nil
}

// PostWriteResponse count write
func (p *Metric) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, e error) error {
	sp := res.ServicePath
	sm := res.ServiceMethod

	if sp == "" {
		return nil
	}

	m := metrics.GetOrRegisterMeter(p.withPrefix("service."+sp+"."+sm+".Write_Qps"), p.Registry)
	m.Mark(1)

	t := ctx.Value(server.StartRequestContextKey).(int64)

	if t > 0 {
		t = time.Now().UnixNano() - t
		if t < 30*time.Minute.Nanoseconds() { //it is impossible that calltime exceeds 30 minute
			//Historgram
			h := metrics.GetOrRegisterHistogram(p.withPrefix("service."+sp+"."+sm+".CallTime"), p.Registry,
				metrics.NewExpDecaySample(1028, 0.015))
			h.Update(t)
		}
	}
	return nil
}

// Log reports metrics into logs.
//
// p.Log( 5 * time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
//
func (p *Metric) Log(freq time.Duration, l metrics.Logger) {
	go metrics.Log(p.Registry, freq, l)
}

// Graphite reports metrics into graphite.
//
// 	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2003")
//  p.Graphite(10e9, "metrics", addr)
//
func (p *Metric) Graphite(freq time.Duration, prefix string, addr *net.TCPAddr) {
	go metrics.Graphite(p.Registry, freq, prefix, addr)
}

// InfluxDB reports metrics into influxdb.
//
// 	p.InfluxDB(10e9, "127.0.0.1:8086","metrics", "test","test"})
//
func (p *Metric) InfluxDB(freq time.Duration, url, database, username, password string) {
	go influxdb.InfluxDB(p.Registry, freq, url, database, username, password)
}
