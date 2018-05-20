package client

import (
	"crypto/tls"
	"frpc/core"
	"frpc/protocol"
	"time"
)

type Option struct {
	SerializeType     protocol.SerializeType
	CompressType      protocol.CompressType
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ConnTimeout       time.Duration
	Heartbeat         bool
	HeartbeatInterval time.Duration
	RPCPath           string
	Retries           int
	failMode          core.FailMode
	Breaker           BreakerClient
	//for quic
	TLSConfig *tls.Config
	//for kcp
	Block interface{}
}

func (c *Client) Serialize(serializeType protocol.SerializeType) *Client {
	c.option.SerializeType = serializeType
	return c
}

func (c *Client) Compress(compressType protocol.CompressType) *Client {
	c.option.CompressType = compressType
	return c
}

func (c *Client) ReadTimeout(timeout time.Duration) *Client {
	c.option.ReadTimeout = timeout
	return c
}

func (c *Client) WriteTimeout(timeout time.Duration) *Client {
	c.option.WriteTimeout = timeout
	return c
}

func (c *Client) ConnTimeout(timeout time.Duration) *Client {
	c.option.ReadTimeout = timeout
	return c
}

func (c *Client) Heartbeat(isHearBeat bool, interval time.Duration) *Client {
	c.option.Heartbeat = isHearBeat
	c.option.HeartbeatInterval = interval
	return c
}

func (c *Client) RpcPath(path string) *Client {
	c.option.RPCPath = path
	return c
}

func (c *Client) Retries(retries int) *Client {
	c.option.Retries = retries
	return c
}

func (c *Client) FailMode(failMode core.FailMode) *Client {
	c.option.failMode = failMode
	return c
}

func (c *Client) WithBlock(kc interface{}) *Client {
	c.option.Block = kc
	return c
}

func (c *Client) WithTls(tls *tls.Config) *Client {
	c.option.TLSConfig = tls
	return c
}
