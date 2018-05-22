package client

import (
	"crypto/tls"
	"frpc/core"
	"frpc/protocol"
	"time"
)

type option struct {
	serializeType     protocol.SerializeType
	compressType      protocol.CompressType
	readTimeout       time.Duration
	writeTimeout      time.Duration
	connTimeout       time.Duration
	heartbeat         bool
	heartbeatInterval time.Duration
	rpcPath           string
	retries           int
	failMode          core.FailMode
	breakerClient     breakerClient
	//for quic
	tlsConfig *tls.Config
	//for kcp
	block interface{}
}

func (c *Client) Serialize(serializeType protocol.SerializeType) *Client {
	c.option.serializeType = serializeType
	return c
}

func (c *Client) Compress(compressType protocol.CompressType) *Client {
	c.option.compressType = compressType
	return c
}

func (c *Client) ReadTimeout(timeout time.Duration) *Client {
	c.option.readTimeout = timeout
	return c
}

func (c *Client) WriteTimeout(timeout time.Duration) *Client {
	c.option.writeTimeout = timeout
	return c
}

func (c *Client) ConnTimeout(timeout time.Duration) *Client {
	c.option.readTimeout = timeout
	return c
}

func (c *Client) Heartbeat(isHearBeat bool, interval time.Duration) *Client {
	c.option.heartbeat = isHearBeat
	c.option.heartbeatInterval = interval
	return c
}

func (c *Client) RpcPath(path string) *Client {
	c.option.rpcPath = path
	return c
}

func (c *Client) Retries(retries int) *Client {
	c.option.retries = retries
	return c
}

func (c *Client) FailMode(failMode core.FailMode) *Client {
	c.option.failMode = failMode
	return c
}

func (c *Client) WithBlock(kc interface{}) *Client {
	c.option.block = kc
	return c
}

func (c *Client) WithTls(tls *tls.Config) *Client {
	c.option.tlsConfig = tls
	return c
}
