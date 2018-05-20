package test

import (
	"context"
	"testing"
	"time"
	"frpc/server"
	"frpc/core"
	"frpc/client"
	"frpc/controller"
	"os"
	"log"
)

func TestMonitorMetricLog(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		Metric(controller.NewMetric().Log(5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))).
		Register(new(Arith))

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Mul", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 200 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Add", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 30 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	select {}
}

func TestMonitorMetricGraphite(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		Metric(controller.NewMetric().CaptureRunTimeStats().Graphite(1e9, "frpc.services.host.127_0_0_1", "127.0.0.1:2003")).
		Register(new(Arith))

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Mul", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 200 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Add", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 30 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	select {}
}

func TestMonitorTrace(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		Trace(controller.NewTrace().ExportListner(":8088")).
		Register(new(Arith))

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Mul", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 200 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	go func() {
		for {
			err := client.CallProxy(context.Background(), "Add", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 30 but got %d", reply.C)
			}

			time.Sleep(time.Duration(500) * time.Microsecond)
		}
	}()

	select {}
}

func TestMonitorRateLimit(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RateLimit(controller.NewConnConcurrentLimit(1000*time.Millisecond, 5)).
		Register(new(Arith))

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	for i := 0; i < 10; i++ {
		go func() {
			client := client.NewClient().
				Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
				Selector(core.Random)

			defer client.Close()

			err := client.CallProxy(context.Background(), "Mul", args, reply)
			if err != nil {
				t.Fatalf("failed to call: %v", err)
			}

			if reply.C != 200 {
				t.Fatalf("expect 200 but got %d", reply.C)
			}

			log.Print(reply.C)
		}()

	}

	select {}

}
