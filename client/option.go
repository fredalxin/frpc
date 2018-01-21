package client

import (
	"time"
	"frpc/protocol"
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
