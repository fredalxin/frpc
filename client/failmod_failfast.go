package client

import (
	"context"
)

type FailFastMod struct{

}

func newFailFastMod() failMod {
	return &FailFastMod{}
}

func (failFastMod FailFastMod) Call(c *Client, client *Client, ctx context.Context,
	servicePath, serviceMethod, cname string, retries int, args interface{}, reply interface{}) error {
	var err error
	err = client.CallDirect(ctx, servicePath, serviceMethod, args, reply)
	if err != nil {
		if _, ok := err.(ServiceError); !ok {
			c.removeClient(cname, client)
		}
	}
	return err
}
