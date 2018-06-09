package client

import (
	"context"
)

type FailOverMod struct{

}

func newFailOverMod() failMod {
	return &FailOverMod{}
}

func (failOverMod FailOverMod) Call(c *Client, client *Client, ctx context.Context,
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
		cname, client, err = c.selectClient(ctx, servicePath, serviceMethod, args)
	}
	return err
}
