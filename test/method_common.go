package test

import (
	"context"
	"time"
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

type Arith2 int

func (t *Arith2) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B * 100
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
