package registry

import (
	"fmt"
	"github.com/cocosip/zero/contrib/registry/local"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
	"sync"
)

type DiscoveryRegistrar interface {
	registry.Discovery
	registry.Registrar
}

type FactoryInterface interface {
	GetRegister() (registry.Registrar, error)
	GetDiscovery() (registry.Discovery, error)
}

type factory struct {
	opt *RegistryOption
	reg DiscoveryRegistrar
	m   *sync.Mutex
}

func New(opt *RegistryOption) FactoryInterface {
	return &factory{
		opt: opt,
		m:   &sync.Mutex{},
	}
}

func (f *factory) GetRegister() (registry.Registrar, error) {
	reg, err := f.getRegistry()
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func (f *factory) GetDiscovery() (registry.Discovery, error) {
	reg, err := f.getRegistry()
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func (f *factory) getRegistry() (DiscoveryRegistrar, error) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.reg != nil {
		return f.reg, nil
	}
	switch strings.ToLower(f.opt.GetProvider()) {
	case "local":
		if f.opt.Local == nil {
			return nil, fmt.Errorf("local registry is nil")
		}
		var entries []*local.ServiceEntry
		for i := range f.opt.Local.Entries {
			e := f.opt.Local.Entries[i]
			entry := &local.ServiceEntry{
				ID:        e.GetId(),
				Name:      e.GetName(),
				Endpoints: e.GetEndpoints(),
				Version:   e.GetVersion(),
			}
			entries = append(entries, entry)
		}
		f.reg = local.New(f.opt.GetAuthority(), entries...)
	case "etcd":
		client, err := clientv3.New(clientv3.Config{
			Endpoints: f.opt.Etcd.GetEndpoints(),
			Username:  f.opt.Etcd.GetUsername(),
			Password:  f.opt.Etcd.GetPassword(),
		})
		if err != nil {
			return nil, err
		}
		f.reg = etcd.New(client)
	}

	if f.reg != nil {
		return f.reg, nil
	}
	return nil, fmt.Errorf("invalid registry %s", f.opt.GetProvider())
}
