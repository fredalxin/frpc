package client

import (
	"context"
)

type FailTryMod struct{

}

func newFailTryMod() failMod {
	return &FailTryMod{}
}

func (failTryMod FailTryMod) Call(c *Client, client *Client, ctx context.Context,
	servicePath, serviceMethod, cname string, retries int, args interface{}, reply interface{}) error {
	var err error
	for retries > 0 {
		retries--
		if client != nil {
			err = client.CallDirect(ctx, servicePath, serviceMethod, args, reply)
			if err == nil {
				return nil
			}
		}
		c.removeClient(cname, client)
		//reconnect
		client, err = c.getCachedClient(cname)
	}
	return err
}
