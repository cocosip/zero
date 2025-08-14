package local

import (
	"context"
	"errors"
	"sync"
	"time"

	kratos_registry "github.com/go-kratos/kratos/v2/registry"
)

// ErrWatcherStopped is returned when the watcher has been stopped.
var ErrWatcherStopped = errors.New("watcher stopped")

// Watcher implements the kratos_registry.Watcher interface for file-based service discovery.
// It monitors changes to the registry file and notifies subscribers of service updates.
type Watcher struct {
	registry    *Registry
	serviceName string
	ctx         context.Context
	cancel      context.CancelFunc
	ch          chan []*kratos_registry.ServiceInstance
	errorCh     chan error
	mu          sync.RWMutex
	stopped     bool
}

// NewWatcher creates a new file-based watcher for the specified service.
// It monitors the registry file for changes and returns updated service instances.
//
// Parameters:
//   - registry: The file registry instance
//   - serviceName: The name of the service to watch
//
// Returns:
//   - registry.Watcher: A new watcher instance
func NewWatcher(registry *Registry, serviceName string) kratos_registry.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		registry:    registry,
		serviceName: serviceName,
		ctx:         ctx,
		cancel:      cancel,
		ch:          make(chan []*kratos_registry.ServiceInstance, 1),
		errorCh:     make(chan error, 1),
	}

	// Start watching in a separate goroutine
	go w.watch()

	return w
}

// Next returns the next set of service instances.
// It blocks until new instances are available or an error occurs.
//
// Returns:
//   - []*registry.ServiceInstance: Updated service instances
//   - error: An error if the watcher is stopped or encounters an issue
func (w *Watcher) Next() ([]*kratos_registry.ServiceInstance, error) {
	w.mu.RLock()
	if w.stopped {
		w.mu.RUnlock()
		return nil, ErrWatcherStopped
	}
	w.mu.RUnlock()

	select {
	case <-w.ctx.Done():
		return nil, ErrWatcherStopped
	case err, ok := <-w.errorCh:
		if !ok {
			return nil, ErrWatcherStopped
		}
		return nil, err
	case instances, ok := <-w.ch:
		if !ok {
			return nil, ErrWatcherStopped
		}
		return instances, nil
	}
}

// Stop stops the watcher and releases associated resources.
// After calling Stop, the watcher should not be used.
//
// Returns:
//   - error: An error if stopping fails
func (w *Watcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return nil
	}

	w.stopped = true
	w.cancel()
	close(w.ch)
	close(w.errorCh)

	return nil
}

// watch monitors the registry file for changes and sends updates to the channel.
// This method runs in a separate goroutine and handles file polling.
func (w *Watcher) watch() {
	ticker := time.NewTicker(time.Second) // Poll every second
	defer ticker.Stop()

	// Send initial state
	if instances, err := w.registry.GetService(w.ctx, w.serviceName); err == nil {
		w.mu.RLock()
		stopped := w.stopped
		w.mu.RUnlock()
		if !stopped {
			select {
			case w.ch <- instances:
			case <-w.ctx.Done():
				return
			}
		}
	}

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// Check if watcher is stopped
			w.mu.RLock()
			stopped := w.stopped
			w.mu.RUnlock()
			if stopped {
				return
			}

			// Check for service changes
			instances, err := w.registry.GetService(w.ctx, w.serviceName)
			if err != nil {
				w.mu.RLock()
				stopped := w.stopped
				w.mu.RUnlock()
				if !stopped {
					select {
					case w.errorCh <- err:
					case <-w.ctx.Done():
						return
					}
				}
				continue
			}

			// Send updated instances
			w.mu.RLock()
			stopped = w.stopped
			w.mu.RUnlock()
			if !stopped {
				select {
				case w.ch <- instances:
				case <-w.ctx.Done():
					return
				}
			}
		}
	}
}