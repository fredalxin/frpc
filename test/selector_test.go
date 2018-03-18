package test

import (
	"testing"
	"time"
	"context"
	"frpc/server"
	"frpc/core"
	"frpc/client"
)

func TestRandomSelector(t *testing.T) {
	initTwoServer()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "Arith", []string{"localhost:32787"}).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			t.Fatalf("failed to call: %v", err)
		}
		println(reply.C)
	}

}

func TestRoundRobinSelector(t *testing.T) {

	initTwoServer()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "Arith", []string{"localhost:32787"}).
		Selector(core.RoundRobin)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			t.Fatalf("failed to call: %v", err)
		}
		println(reply.C)
	}

}


func TestHashSelector(t *testing.T) {

	initTwoServer()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "Arith", []string{"localhost:32787"}).
		Selector(core.Hash)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			t.Fatalf("failed to call: %v", err)
		}
		println(reply.C)
	}

}

func initTwoServer() {
	s1, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(Arith), "Arith", "")

	go s1.ServeProxy()

	defer s1.Close()

	time.Sleep(500 * time.Millisecond)


	s2, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8974", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(Arith2), "Arith", "")

	go s2.ServeProxy()

	defer s2.Close()

	time.Sleep(500 * time.Millisecond)
}
