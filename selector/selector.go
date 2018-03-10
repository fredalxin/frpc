package selector

import (
	"context"
	"frpc/core"
)

type Selector interface {
	Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string
	UpdateServer(servers map[string]string)
}

func NewSelector(selectMode core.SelectMode, servers map[string]string) Selector {
	switch selectMode {
	case core.Random:
		return newRandomSelector(servers)
	case core.RoundRobin:
		return newRoundRobinSelector(servers)
	case core.Weighted:
		return newWeightedSelector(servers)
	case core.Hash:
		return newHashSelector(servers)
	case core.Ping:
		return newPingSelector(servers)
	default:
		return newRandomSelector(servers)
	}
}
