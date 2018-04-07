package test

import (
	"context"
	"time"
	"errors"
	"fmt"
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

func (t *Arith) Add(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type Arith2 int

func (t *Arith2) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B * 100
	return nil
}

type ArithF int

func (t *ArithF) Mul(ctx context.Context, args *Args, reply *Reply) error {
	return errors.New("unknown error")
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

var count = -1

type ArithB int

func (t *ArithB) Mul(ctx context.Context, args *Args, reply *Reply) error {
	count++
	if count >= 5 && count < 10 {
		return nil
	}
	return errors.New("test error")
}
