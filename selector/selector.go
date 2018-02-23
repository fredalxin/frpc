package selector

import (
	"context"
)

type Selector interface {
	Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string
	UpdateServer(servers map[string]string)
}

func NewSelector(selectMode string, servers map[string]string) Selector {
	switch selectMode {
	case "random":
		return newRandomSelector(servers)
	case "roundrobin":
		return newRoundRobinSelector(servers)
	case "weight":
		return newWeightedSelector(servers)
	case "hash":
		return newHashSelector(servers)
	case "ping":
		return newPingSelector(servers)
	//case "geo":
	//	return newGeoSelector(servers)
	default:
		return newRandomSelector(servers)
	}
}
