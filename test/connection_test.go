package test

import (
	"testing"
	"time"
	"frpc/server"
	"golang.org/x/crypto/pbkdf2"
	"crypto/sha1"
	"github.com/xtaci/kcp-go"
	"frpc/core"
	"frpc/client"
	"context"
	"crypto/tls"
	"frpc/log"
)

const cryptKey = "frpc-key"
const cryptSalt = "frpc-salt"

func TestKcp(t *testing.T) {
	key := pbkdf2.Key([]byte(cryptKey), []byte(cryptSalt), 4096, 32, sha1.New)
	bc, err := kcp.NewAESBlockCrypt(key)
	if err != nil {
		panic(err)
	}
	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "kcp@localhost:8972", []string{"localhost:32787"}, time.Minute).
		WithBlockCrypt(bc).
		Register(new(Arith))

	go s.ServeProxy()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)

	client := client.NewClient().
		Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
		WithBlock(bc).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	err = client.CallProxy(context.Background(), "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}

	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

	println(reply.C)

}

func TestQuic(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
	if err != nil {
		log.Error(err)
		return
	}
	serverConf := &tls.Config{Certificates: []tls.Certificate{cert}}

	s, _ := server.
		NewServer().
		Registry(core.Consul, "/frpc_test", "quic@localhost:8972", []string{"localhost:32787"}, time.Minute).
		WithTlsConfig(serverConf).
		Register(new(Arith))

	go s.ServeProxy()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)

	clientConf := &tls.Config{
		InsecureSkipVerify: true,
	}
	client := client.NewClient().
		Discovery(core.Consul, "/frpc_test", "Arith", []string{"localhost:32787"}).
		WithTls(clientConf).
		Selector(core.Random)

	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	err = client.CallProxy(context.Background(), "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}

	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

	println(reply.C)

}
