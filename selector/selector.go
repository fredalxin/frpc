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
	case "RandomSelect":
		return newRandomSelector(servers)
	default:
		return newRandomSelector(servers)
	}
}
