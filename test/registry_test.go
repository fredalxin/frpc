package test

import (
	"context"
	"frpc/client"
	"frpc/core"
	"frpc/server"
	"testing"
	"time"
)

func TestETCD(t *testing.T) {
	s := server.NewServer().
		Registry(core.Etcd, "/frpc_test", "tcp@localhost:8972", []string{"localhost:2379"}, time.Minute)

	s.Register(new(Arith))
	go s.Serve("tcp", "127.0.0.1:8972")
	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Etcd, "/frpc_test", "Arith", []string{"localhost:2379"}).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	err := client.CallProxy(context.Background(), "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}

	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

	println(reply.C)
}

func TestCONSUL(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
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
	err := client.CallProxy(context.Background(), "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}

	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

	println(reply.C)
}
