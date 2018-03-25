package test

import (
	"time"
	"frpc/server"
	"frpc/core"
	"context"
	"testing"
	"frpc/client"
	"log"
)

func TestFailTry(t *testing.T) {
	s1, s2 := initTwoFailServer()
	defer s1.Close()
	defer s2.Close()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithF", []string{"localhost:32787"}).
		Selector(core.Random).
		Retries(3).
		FailMode(core.FailTry)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Printf("failed to call: %v", err.Error())
		} else {
			log.Printf("%d * %d = %d", args.A, args.B, reply.C)
		}
	}
}

func TestFailOver(t *testing.T) {
	s1, s2 := initTwoFailServer()
	defer s1.Close()
	defer s2.Close()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithF", []string{"localhost:32787"}).
		Selector(core.Random).
		Retries(10).
		FailMode(core.FailOver)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Printf("failed to call: %v", err.Error())
		} else {
			log.Printf("%d * %d = %d", args.A, args.B, reply.C)
		}
	}
}

func TestFailFast(t *testing.T) {
	s1, s2 := initTwoFailServer()
	defer s1.Close()
	defer s2.Close()

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithF", []string{"localhost:32787"}).
		Selector(core.RoundRobin).
		Retries(10).
		FailMode(core.FailFast)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	for i := 0; i < 10; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Printf("failed to call: %v", err)
		} else {
			log.Printf("%d * %d = %d", args.A, args.B, reply.C)
		}
	}
}

func initTwoFailServer() (*server.Server, *server.Server) {
	s1, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(ArithF), "ArithF", "")

	go s1.ServeProxy()

	time.Sleep(500 * time.Millisecond)

	s2, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8973", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(Arith), "ArithF", "")

	go s2.ServeProxy()

	time.Sleep(500 * time.Millisecond)
	return s1, s2
}
