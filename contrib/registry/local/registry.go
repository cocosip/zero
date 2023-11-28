package local

import (
	"context"
	"github.com/go-kratos/kratos/v2/registry"
	"slices"
	"strings"
	"sync"
)

var (
	_ registry.Registrar = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

type ServiceEntry struct {
	ID        string
	Name      string
	Endpoints []string
	Version   string
}

func NewServiceEntry(id, name, version string, endpoints ...string) *ServiceEntry {
	if strings.TrimSpace(id) == "" {
		id = name
	}
	return &ServiceEntry{
		ID:        id,
		Name:      name,
		Endpoints: endpoints,
		Version:   version,
	}
}

type Registry struct {
	entries map[string]*ServiceEntry
	m       *sync.Mutex
}

func New(entries ...*ServiceEntry) *Registry {
	r := &Registry{
		m: &sync.Mutex{},
	}

	for _, entry := range entries {
		r.entries[entry.Name] = entry
	}
	return r
}

func (r *Registry) Register(_ context.Context, service *registry.ServiceInstance) error {
	r.m.Lock()
	defer r.m.Unlock()
	if entry, ok := r.entries[service.Name]; ok {
		for _, endpoint := range service.Endpoints {
			if !slices.Contains(entry.Endpoints, endpoint) {
				entry.Endpoints = append(entry.Endpoints, endpoint)
			}
		}
		return nil
	}

	entry := NewServiceEntry(service.ID, service.Name, service.Version, service.Endpoints...)
	r.entries[entry.Name] = entry
	return nil
}

func (r *Registry) Deregister(_ context.Context, service *registry.ServiceInstance) error {
	r.m.Lock()
	defer r.m.Unlock()
	if entry, ok := r.entries[service.Name]; ok {
		if entry.Name == service.Name && entry.ID == service.ID {
			delete(r.entries, service.Name)
		}
	}
	return nil
}

func (r *Registry) GetService(_ context.Context, name string) ([]*registry.ServiceInstance, error) {
	r.m.Lock()
	defer r.m.Unlock()
	items := make([]*registry.ServiceInstance, 0)
	if entry, ok := r.entries[name]; ok {
		item := &registry.ServiceInstance{
			ID:        entry.ID,
			Name:      entry.Name,
			Version:   entry.Version,
			Metadata:  make(map[string]string),
			Endpoints: entry.Endpoints,
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Registry) Watch(_ context.Context, name string) (registry.Watcher, error) {
	return newWatcher(name)
}
