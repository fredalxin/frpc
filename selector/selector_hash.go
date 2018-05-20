package selector

import "context"

type hashSelector struct {
	servers []string
}

func newHashSelector(servers map[string]string) Selector {
	var ss = make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	return &hashSelector{servers: ss}
}

func (s hashSelector) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	ss := s.servers
	if len(ss) == 0 {
		return ""
	}
	i := JumpConsistentHash(len(ss), servicePath, serviceMethod, args)
	return ss[i]
}

func (s *hashSelector) UpdateServer(servers map[string]string) {
	var ss = make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	s.servers = ss
}
