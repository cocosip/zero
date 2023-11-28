package local

import "github.com/go-kratos/kratos/v2/registry"

var _ registry.Watcher = (*watcher)(nil)

type watcher struct {
	name string
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	return make([]*registry.ServiceInstance, 0), nil
}

func (w *watcher) Stop() error {
	return nil
}

func newWatcher(name string) (*watcher, error) {
	return &watcher{name: name}, nil
}
