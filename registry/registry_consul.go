package registry

import (
	"errors"
	"fmt"
	"frpc/log"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"net/url"
	"strings"
	"sync"
	"time"
)

func init() {
	consul.Register()
}

type ConsulRegister struct {
	// service address, for example, tcp@127.0.0.1:8972, quic@127.0.0.1:1234
	ServiceAddress string
	// etcd addresses
	ConsulAddr []string
	// base path for frpc server, for example com/example/frpc
	BasePath string
	// Registered services
	Services       []string
	metasLock      sync.RWMutex
	metas          map[string]string
	UpdateInterval time.Duration

	Options *store.Config
	store   store.Store
}

func NewConsulRegistry(basePath string, serviceAddress string, consulAddr []string, interval time.Duration) Registry {
	return &ConsulRegister{
		ServiceAddress: serviceAddress,
		ConsulAddr:     consulAddr,
		BasePath:       basePath,
		UpdateInterval: interval,
	}
}

func (p *ConsulRegister) Start() error {
	if p.store == nil {
		store, err := libkv.NewStore(store.CONSUL, p.ConsulAddr, p.Options)
		if err != nil {
			log.Errorf("cannot create etcd store: %v", err)
			return err
		}
		p.store = store
	}

	if p.BasePath[0] == '/' {
		p.BasePath = p.BasePath[1:]
	}

	err := p.store.Put(p.BasePath, []byte("frpc_path"), &store.WriteOptions{IsDir: true})
	if err != nil {
		log.Errorf("cannot create etcd path %s: %v", p.BasePath, err)
		return err
	}

	if p.UpdateInterval > 0 {
		ticker := time.NewTicker(p.UpdateInterval)
		go func() {
			defer p.store.Close()

			for range ticker.C {
				var data []byte
				for _, name := range p.Services {
					nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
					kvPair, err := p.store.Get(nodePath)
					if err != nil {
						log.Infof("can't get data of node: %s, because of %v", nodePath, err.Error())

						p.metasLock.RLock()
						meta := p.metas[name]
						p.metasLock.RUnlock()

						err = p.store.Put(nodePath, []byte(meta), &store.WriteOptions{TTL: p.UpdateInterval * 3})
						if err != nil {
							log.Errorf("cannot re-create etcd path %s: %v", nodePath, err)
						}

					} else {
						v, _ := url.ParseQuery(string(kvPair.Value))
						v.Set("tps", string(data))
						p.store.Put(nodePath, []byte(v.Encode()), &store.WriteOptions{TTL: p.UpdateInterval * 3})
					}
				}

			}
		}()
	}

	return nil
}

func (p *ConsulRegister) Register(name string, rcvr interface{}, metadata string) (err error) {
	if "" == strings.TrimSpace(name) {
		err = errors.New("Register service `name` can't be empty")
		return
	}

	if p.store == nil {
		consul.Register()
		store, err := libkv.NewStore(store.CONSUL, p.ConsulAddr, p.Options)
		if err != nil {
			log.Errorf("cannot create etcd registry: %v", err)
			return err
		}
		p.store = store
	}

	err = p.store.Put(p.BasePath, []byte("frpc_path"), &store.WriteOptions{IsDir: true})
	if err != nil {
		log.Errorf("cannot create etcd path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.store.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	err = p.store.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: p.UpdateInterval * 2})
	if err != nil {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	p.Services = append(p.Services, name)

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	p.metas[name] = metadata
	p.metasLock.Unlock()
	return
}
