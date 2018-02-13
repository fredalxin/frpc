package client

import (
	"context"
	"testing"
	"time"
	"frpc/server"
	"frpc/protocol"
	"frpc/core"
	"fmt"
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

type PBArith int

func (t *PBArith) Mul(ctx context.Context, args *ProtoArgs, reply *ProtoReply) error {
	reply.C = args.A * args.B
	return nil
}

type TimeoutArith int

func (t *TimeoutArith) Mul(ctx context.Context, args *ProtoArgs, reply *ProtoReply) error {
	time.Sleep(2 * time.Second)
	reply.C = args.A * args.B
	return nil
}

type MetaDataArith int

func (t *MetaDataArith) Mul(ctx context.Context, args *ProtoArgs, reply *ProtoReply) error {
	reqMetaData := ctx.Value(core.ReqMetaDataKey).(map[string]string)
	fmt.Println("server received meta:", reqMetaData)
	respMetaData := ctx.Value(core.ResMetaDataKey).(map[string]string)
	respMetaData["echo"] = "from server"
	reply.C = args.A * args.B
	return nil
}

func TestClient(t *testing.T) {
	s := initServer()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := initClient(t)
	defer c.Close()
	args, reply := initParam()
	doCall(t, c, args, reply)
}

func TestAsyncClient(t *testing.T) {
	s := initServer()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := initClient(t)
	defer c.Close()
	args, reply := initParam()
	doAsyncCall(t, c, args, reply)
}

func TestHttpClient(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(Arith), "Arith")
	go s.ServePath("http", "127.0.0.1:8080", "testPath")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := NewClient().RpcPath("testPath")
	err := c.Connect("http", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()
	args, reply := initParam()
	doCall(t, c, args, reply)

}

func TestCodecAndCompressClient(t *testing.T) {
	s := initServer()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := initClient(t)
	defer c.Close()
	args, reply := initParam()
	//json
	c.Serialize(protocol.JSON)
	doCall(t, c, args, reply)
	//msgpack
	c.Serialize(protocol.MsgPack)
	doCall(t, c, args, reply)
	//gzip
	c.Compress(protocol.Gzip)
	doCall(t, c, args, reply)
}

func TestProtobuf(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(PBArith), "PBArith")
	go s.Serve("tcp", "127.0.0.1:8080")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := initClient(t).Serialize(protocol.ProtoBuffer)
	defer c.Close()
	pbArgs := &ProtoArgs{
		A: 10,
		B: 20,
	}
	pbReply := &ProtoReply{
	}
	doCallProto(t, c, pbArgs, pbReply)
}

func TestTimeout(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(TimeoutArith), "TimeoutArith")
	go s.Serve("tcp", "127.0.0.1:8080")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := NewClient()
	c.ReadTimeout(1 * time.Second)
	err := c.Connect("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()
	args, reply := initParam()
	doCallPath(t, c, "TimeoutArith", args, reply)
}

func TestMetaData(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(MetaDataArith), "MetaDataArith")
	go s.Serve("tcp", "127.0.0.1:8080")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := NewClient()
	err := c.Connect("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()
	args, reply := initParam()
	ctx := context.
		WithValue(context.Background(), core.ReqMetaDataKey, map[string]string{"aaa": "from client"})
	ctx = context.WithValue(ctx, core.ResMetaDataKey, make(map[string]string))

	err = c.Call(ctx, "MetaDataArith", "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

	resMetaData := ctx.Value(core.ResMetaDataKey).(map[string]string)
	fmt.Println("client received meta:", resMetaData)
}

func TestHeartBeat(t *testing.T) {
	s := server.NewServer()
	s.RegisterName(new(Arith), "Arith")
	go s.Serve("tcp", "127.0.0.1:8080")
	defer s.Close()
	time.Sleep(500 * time.Millisecond)
	c := NewClient().Heartbeat(true, time.Second)
	err := c.Connect("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	defer c.Close()

	args, reply := initParam()
	err = c.Call(context.Background(), "Arith", "Mul", args, reply)
	println(reply.C)

	time.Sleep(10*time.Minute)
}

func initServer() *server.Server {
	s := server.NewServer()
	s.RegisterName(new(Arith), "Arith")
	go s.Serve("tcp", "127.0.0.1:8080")
	return s
}

func initClient(t *testing.T) *Client {
	c := NewClient()
	err := c.Connect("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to connect:v%", err)
	}
	return c
}

func initParam() (*Args, *Reply) {
	args := &Args{
		A: 10,
		B: 20,
	}
	reply := &Reply{}
	return args, reply
}

func doCall(t *testing.T, c *Client, args interface{}, reply *Reply) {
	doCallPath(t, c, "Arith", args, reply)
}

func doAsyncCall(t *testing.T, c *Client, args interface{}, reply *Reply) {
	call := c.Go(context.Background(), "Arith", "Mul", args, reply, nil)
	rC := <-call.Done
	if rC.Error != nil {
		t.Fatalf("failed to call: %v", rC.Error)
	} else {
		if reply.C != 200 {
			t.Fatalf("expect 200 but got %d", reply.C)
		}
		println(reply.C)
	}
}

func doCallPath(t *testing.T, c *Client, path string, args interface{}, reply *Reply) {
	err := c.Call(context.Background(), path, "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}
	println(reply.C)
}

func doCallProto(t *testing.T, c *Client, args *ProtoArgs, reply *ProtoReply) {
	err := c.Call(context.Background(), "PBArith", "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}
	println(reply.C)
}
