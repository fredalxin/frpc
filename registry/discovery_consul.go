package registry

import (
	"sync"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

func init() {
	etcd.Register()
}

type ConsulDiscovery struct {
	basePath string
	store    store.Store
	pairs    []*KV
	chans    []chan []*KV
	mu       sync.Mutex

	// -1 means it always retry to watch until watch is ok, 0 means no retry.
	//需可设置
	RetriesAfterWatchFailed int

	stopCh chan struct{}
}

func NewConsulDiscovery(basePath string, servicePath string, etcdAddr []string) Discovery {
	return nil
}
