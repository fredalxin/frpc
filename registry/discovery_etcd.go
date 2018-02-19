package registry

import (
	"sync"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv"
	"frpc/log"
	"strings"
	"time"
	"github.com/docker/libkv/store/etcd"
)

func init() {
	etcd.Register()
}

type EtcdDiscovery struct {
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

func NewEtcdDiscovery(basePath string, servicePath string, etcdAddr []string) Discovery {
	return NewEtcdDiscoveryOption(basePath, servicePath, etcdAddr, nil)
}

func NewEtcdDiscoveryOption(basePath string, servicePath string, etcdAddr []string, options *store.Config) Discovery {
	store, err := libkv.NewStore(store.ETCD, etcdAddr, options)
	if err != nil {
		log.Infof("cannot create etcd store: %v", err)
		panic(err)
	}

	return newEtcdDiscoveryStore(basePath+"/"+servicePath, store)
}

func newEtcdDiscoveryStore(basePath string, store store.Store) Discovery {
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	d := &EtcdDiscovery{basePath: basePath, store: store}
	d.stopCh = make(chan struct{})

	kv, err := store.List(basePath)
	if err != nil {
		log.Infof("cannot get services of from registry: %v", basePath, err)
		panic(err)
	}
	var pairs = make([]*KV, 0, len(kv))
	prefix := d.basePath + "/"
	for _, entry := range kv {
		k := strings.TrimPrefix(entry.Key, prefix)
		pairs = append(pairs, &KV{Key: k, Value: string(entry.Value)})
	}
	d.pairs = pairs
	d.RetriesAfterWatchFailed = -1

	go d.watch()
	return d
}

func (d *EtcdDiscovery) watch() {
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

	readChanges:
		for {
			select {
			case <-d.stopCh:
				log.Info("discovery has been closed")
				return
			case kv := <-c:
				if kv == nil {
					break readChanges
				}
				var pairs []*KV // latest servers
				prefix := d.basePath + "/"
				for _, entry := range kv {
					k := strings.TrimPrefix(entry.Key, prefix)
					pairs = append(pairs, &KV{Key: k, Value: string(entry.Value)})
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
func (d EtcdDiscovery) Clone(servicePath string) Discovery {
	return newEtcdDiscoveryStore(d.basePath+"/"+servicePath, d.store)
}

// GetServices returns the servers
func (d EtcdDiscovery) GetServices() []*KV {
	return d.pairs
}

// WatchService returns a nil chan.
func (d *EtcdDiscovery) WatchService() chan []*KV {
	ch := make(chan []*KV, 10)
	d.chans = append(d.chans, ch)
	return ch
}

func (d *EtcdDiscovery) RemoveWatcher(ch chan []*KV) {
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

func (d *EtcdDiscovery) Close() {
	close(d.stopCh)
}
