package controller

import (
	"net"
	"time"
	"github.com/juju/ratelimit"
	"context"
	"frpc/protocol"
	"frpc/log"
)

type RateLimit struct {
	FillInterval time.Duration
	Capacity     int64
	bucket       *ratelimit.Bucket
}

func NewConnConcurrentLimit(fillInterval time.Duration, capacity int64) *RateLimit {
	tb := ratelimit.NewBucket(fillInterval, capacity)

	return &RateLimit{
		FillInterval: fillInterval,
		Capacity:     capacity,
		bucket:       tb,
	}
}

func (r *RateLimit) Register(name string, rcvr interface{}, metadata string) error {
	return nil
}

func (r *RateLimit) HandleConn(conn net.Conn) (net.Conn, bool) {
	i := r.bucket.TakeAvailable(1)
	if i > 0 {
		return conn, true
	} else {
		log.Error("conn out of limit")
		return conn, false
	}
}

func (r *RateLimit) PostRequest(ctx context.Context, req *protocol.Message, err error) error {
	//todo
	return nil
}

func (r *RateLimit) PostResponse(ctx context.Context, req *protocol.Message, res *protocol.Message, err error) error {
	//todo
	return nil
}
