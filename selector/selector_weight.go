package selector

import (
	"net/url"
	"strconv"
	"context"
)
type weightedSelector struct {
	servers []*Weighted
}

func newWeightedSelector(servers map[string]string) Selector {
	ss := createWeighted(servers)
	return &weightedSelector{servers: ss}
}

func (s *weightedSelector) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	ss := s.servers
	if len(ss) == 0 {
		return ""
	}
	w := nextWeighted(ss)
	if w == nil {
		return ""
	}
	return w.Server
}

func (s *weightedSelector) UpdateServer(servers map[string]string) {
	ss := createWeighted(servers)
	s.servers = ss
}

func createWeighted(servers map[string]string) []*Weighted {
	var ss = make([]*Weighted, 0, len(servers))
	for k, metadata := range servers {
		w := &Weighted{Server: k, Weight: 1, EffectiveWeight: 1}

		if v, err := url.ParseQuery(metadata); err == nil {
			ww := v.Get("weight")
			if ww != "" {
				if weight, err := strconv.Atoi(ww); err == nil {
					w.Weight = weight
					w.EffectiveWeight = weight
				}
			}
		}

		ss = append(ss, w)
	}

	return ss
}
