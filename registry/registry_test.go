package registry_test

import (
	"testing"
	"time"
	"frpc/server"
	"context"
	"frpc/client"
	"frpc/core"
)

type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func TestETCD(t *testing.T) {
	s := server.NewServer().
		Registry(core.Etcd, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:2379"}, time.Minute)

	s.RegisterName(new(Arith), "Arith", "")
	go s.Serve("tcp", "127.0.0.1:8972")
	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Etcd, "/rpcx_test", "Arith", []string{"localhost:2379"}).
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
	s := server.NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32780"}, time.Minute)

	s.RegisterName(new(Arith), "Arith", "")
	go s.Serve("tcp", "127.0.0.1:8972")
	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "Arith", []string{"localhost:32780"}).
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
