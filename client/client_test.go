package client

import (
	"context"
	"testing"
	"time"
	"frpc/server"
	"frpc/protocol"
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

func TestCodecClient(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(Arith), "Arith")
	go s.Serve("tcp", "127.0.0.1:8080")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)

	c := NewClient(DefaultOption)
	err := c.Connect("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()

	args := &Args{
		A: 10,
		B: 20,
	}
	reply := &Reply{}
	c.option.SerializeType = protocol.JSON
	doCall(t, c, args, reply)
	c.option.SerializeType = protocol.MsgPack
	doCall(t, c, args, reply)

	pbArgs := &ProtoArgs{
		A: 10,
		B: 20,
	}
	pbReply := &ProtoReply{
	}

	doCallProto(t, c, pbArgs, pbReply)
}

func TestHttpClient(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(Arith), "Arith")
	go s.ServePath("http", "127.0.0.1:8080", "testPath")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)

	c := NewClient(DefaultOption)
	c.option.RPCPath = "testPath"
	err := c.Connect("http", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()

	args := &Args{
		A: 10,
		B: 20,
	}
	reply := &Reply{
	}

	doCall(t, c, args, reply)

}

func doCall(t *testing.T, c *Client, args interface{}, reply *Reply) {
	err := c.Call(context.Background(), "Arith", "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}
	println(reply.C)
}

func doCallProto(t *testing.T, c *Client, args interface{}, reply *ProtoReply) {
	err := c.Call(context.Background(), "Arith", "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}
	println(reply.C)
}
