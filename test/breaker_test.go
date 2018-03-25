package test

import (
	"context"
	"testing"
	"time"
	"frpc/server"
	"frpc/core"
	"frpc/client"
	"log"
	"frpc/circuit"
	"errors"
)

func TestLocalCircuitBreaker(t *testing.T) {
	count := -1
	fn := func() error {
		count++
		if count >= 5 && count < 10 {
			return nil
		}

		return errors.New("test error")
	}

	cb := circuit.NewSimpleCircuitBreaker(5, 100*time.Millisecond)

	for i := 0; i < 25; i++ {
		err := cb.Call(fn, 200*time.Millisecond)
		log.Printf("got err: %v", err)
		if i == 9 { // expired
			time.Sleep(150 * time.Millisecond)
		}
	}

}

func TestSimpleBreaker(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(ArithB), "ArithB", "")

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithB", []string{"localhost:32787"}).
		Selector(core.Random).
		FailMode(core.FailFast).
		Breaker(circuit.NewSimpleCircuitBreaker(5, 100*time.Millisecond),0)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	for i := 0; i < 25; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err == nil {
			err = errors.New("success")
		}
		log.Printf("got err: %v", err)
		if i == 9 { // expired
			time.Sleep(150 * time.Millisecond)
		}
	}
}


func TestThresholdBreaker(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(ArithB), "ArithB", "")

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithB", []string{"localhost:32787"}).
		Selector(core.Random).
		FailMode(core.FailFast).
		Breaker(circuit.NewThresholdBreaker(5),0)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	for i := 0; i < 25; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err == nil {
			err = errors.New("success")
		}
		log.Printf("got err: %v", err)
		if i == 9 { // expired
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func TestConsecutiveBreaker(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(ArithB), "ArithB", "")

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithB", []string{"localhost:32787"}).
		Selector(core.Random).
		FailMode(core.FailFast).
		Breaker(circuit.NewConsecutiveBreaker(5),0)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	for i := 0; i < 25; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err == nil {
			err = errors.New("success")
		}
		log.Printf("got err: %v", err)
		if i == 9 { // expired
			time.Sleep(150 * time.Millisecond)
		}
	}
}


func TestRateBreaker(t *testing.T) {
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/rpcx_test", "tcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		RegisterName(new(ArithB), "ArithB", "")

	go s.ServeProxy()

	defer s.Close()

	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/rpcx_test", "ArithB", []string{"localhost:32787"}).
		Selector(core.Random).
		FailMode(core.FailFast).
		Breaker(circuit.NewRateBreaker(0.5,10),0)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}

	for i := 0; i < 25; i++ {
		err := client.CallProxy(context.Background(), "Mul", args, reply)
		if err == nil {
			err = errors.New("success")
		}
		log.Printf("got err: %v", err)
		if i == 9 { // expired
			time.Sleep(150 * time.Millisecond)
		}
	}
}
