package registry

import (
	"sync"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv"
	"frpc/log"
	"strings"
	"time"
)

func init() {
	consul.Register()
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
	return NewConsulDiscoveryOption(basePath, servicePath, etcdAddr, nil)
}

func NewConsulDiscoveryOption(basePath string, servicePath string, etcdAddr []string, options *store.Config) Discovery {
	store, err := libkv.NewStore(store.ETCD, etcdAddr, options)
	if err != nil {
		log.Infof("cannot create consul store: %v", err)
		panic(err)
	}

	return newConsulDiscoveryStore(basePath+"/"+servicePath, store)
}

// NewConsulDiscoveryStore returns a new ConsulDiscovery with specified store.
func newConsulDiscoveryStore(basePath string, store store.Store) Discovery {
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	d := &ConsulDiscovery{basePath: basePath, store: store}
	d.stopCh = make(chan struct{})

	ps, err := store.List(basePath)
	if err != nil {
		log.Infof("cannot get services of from registry: %v", basePath, err)
		panic(err)
	}

	var pairs = make([]*KV, 0, len(ps))
	prefix := d.basePath + "/"
	for _, p := range ps {
		k := strings.TrimPrefix(p.Key, prefix)
		pairs = append(pairs, &KV{Key: k, Value: string(p.Value)})
	}
	d.pairs = pairs
	d.RetriesAfterWatchFailed = -1
	go d.watch()
	return d
}

func (d *ConsulDiscovery) watch() {
	for {
		var err error
		var c <-chan []*store.KVPair
		var tempDelay time.Duration

		retry := d.RetriesAfterWatchFailed
		for d.RetriesAfterWatchFailed == -1 || retry > 0 {
			c, err = d.store.WatchTree(d.basePath, nil)
			if err != nil {
				if d.RetriesAfterWatchFailed > 0 {
					retry--
				}
				if tempDelay == 0 {
					tempDelay = 1 * time.Second
				} else {
					tempDelay *= 2
				}
				if max := 30 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Warnf("can not watchtree (with retry %d, sleep %v): %s: %v", retry, tempDelay, d.basePath, err)
				time.Sleep(tempDelay)
				continue
			}
			break
		}

		if err != nil {
			log.Errorf("can't watch %s: %v", d.basePath, err)
			return
		}

		prefix := d.basePath + "/"

	readChanges:
		for {
			select {
			case <-d.stopCh:
				log.Info("discovery has been closed")
				return
			case ps := <-c:
				if ps == nil {
					break readChanges
				}
				var pairs []*KV // latest servers
				for _, p := range ps {
					k := strings.TrimPrefix(p.Key, prefix)
					pairs = append(pairs, &KV{Key: k, Value: string(p.Value)})
				}
				d.pairs = pairs

				for _, ch := range d.chans {
					ch := ch
					go func() {
						defer func() {
							if r := recover(); r != nil {

							}
						}()
						select {
						case ch <- pairs:
						case <-time.After(time.Minute):
							log.Warn("chan is full and new change has been dropped")
						}
					}()
				}
			}
		}

		log.Warn("chan is closed and will rewatch")
	}
}

// Clone clones this ServiceDiscovery with new servicePath.
func (d ConsulDiscovery) Clone(servicePath string) Discovery {
	return newConsulDiscoveryStore(d.basePath+"/"+servicePath, d.store)
}

// GetServices returns the servers
func (d ConsulDiscovery) GetServices() []*KV {
	return d.pairs
}

// WatchService returns a nil chan.
func (d *ConsulDiscovery) WatchService() chan []*KV {
	ch := make(chan []*KV, 10)
	d.chans = append(d.chans, ch)
	return ch
}

func (d *ConsulDiscovery) RemoveWatcher(ch chan []*KV) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var chans []chan []*KV
	for _, c := range d.chans {
		if c == ch {
			continue
		}

		chans = append(chans, c)
	}

	d.chans = chans
}

func (d *ConsulDiscovery) Close() {
	close(d.stopCh)
}
