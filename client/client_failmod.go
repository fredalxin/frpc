package client

import (
	"context"
	"frpc/core"
)

type failMod interface {
	Call(c *Client, client *Client, ctx context.Context,
		servicePath, serviceMethod, cname string, retries int, args interface{}, reply interface{}) error
}

func newFailMod(failMode core.FailMode) failMod {
	switch failMode {
	case core.FailFast:
		return newFailFastMod()
	case core.FailOver:
		return newFailOverMod()
	case core.FailTry:
		return newFailTryMod()
	default:
		return newFailFastMod()
	}
}
