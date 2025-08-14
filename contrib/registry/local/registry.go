package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	kratos_registry "github.com/go-kratos/kratos/v2/registry"
)

// Registry implements the Kratos kratos_registry.Registrar and kratos_registry.Discovery interfaces
// using local file storage for service registration and discovery.
// This implementation is suitable for scenarios where traditional service discovery
// components are not available, particularly on Windows machines.
type Registry struct {
	filePath string
	mu       sync.RWMutex
	watchers map[string]*Watcher
}

// ServiceInstance represents a service instance stored in the registry file
type ServiceInstance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []string          `json:"endpoints"`
	Timestamp int64             `json:"timestamp"`
}

// RegistryData represents the structure of the registry file
type RegistryData struct {
	Services map[string][]*ServiceInstance `json:"services"`
	Version  string                       `json:"version"`
	Updated  int64                        `json:"updated"`
}

// New creates a new file-based registry instance.
//
// Parameters:
//   - filePath: The path to the registry file where service instances will be stored
//
// Returns:
//   - *Registry: A new registry instance
//   - error: An error if the registry cannot be initialized
func New(filePath string) (*Registry, error) {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Initialize the registry file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		initialData := &RegistryData{
			Services: make(map[string][]*ServiceInstance),
			Version:  "1.0.0",
			Updated:  time.Now().Unix(),
		}
		if err := writeRegistryFile(filePath, initialData); err != nil {
			return nil, fmt.Errorf("failed to initialize registry file: %w", err)
		}
	}

	return &Registry{
		filePath: filePath,
		watchers: make(map[string]*Watcher),
	}, nil
}

// Register registers a service instance to the registry.
//
// Parameters:
//   - ctx: The context for the operation
//   - service: The service instance to register
//
// Returns:
//   - error: An error if the registration fails
func (r *Registry) Register(ctx context.Context, service *kratos_registry.ServiceInstance) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.readRegistryFile()
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	// Convert Kratos ServiceInstance to our internal format
	instance := &ServiceInstance{
		ID:        service.ID,
		Name:      service.Name,
		Version:   service.Version,
		Metadata:  service.Metadata,
		Endpoints: service.Endpoints,
		Timestamp: time.Now().Unix(),
	}

	// Add or update the service instance
	if data.Services == nil {
		data.Services = make(map[string][]*ServiceInstance)
	}

	services := data.Services[service.Name]
	found := false
	for i, existing := range services {
		if existing.ID == service.ID {
			services[i] = instance
			found = true
			break
		}
	}

	if !found {
		services = append(services, instance)
	}

	data.Services[service.Name] = services
	data.Updated = time.Now().Unix()

	if err := writeRegistryFile(r.filePath, data); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	// Notify watchers
	r.notifyWatchers(service.Name)

	return nil
}

// Deregister removes a service instance from the registry.
//
// Parameters:
//   - ctx: The context for the operation
//   - service: The service instance to deregister
//
// Returns:
//   - error: An error if the deregistration fails
func (r *Registry) Deregister(ctx context.Context, service *kratos_registry.ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.readRegistryFile()
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	services := data.Services[service.Name]
	for i, existing := range services {
		if existing.ID == service.ID {
			// Remove the service instance
			services = append(services[:i], services[i+1:]...)
			break
		}
	}

	if len(services) == 0 {
		delete(data.Services, service.Name)
	} else {
		data.Services[service.Name] = services
	}

	data.Updated = time.Now().Unix()

	if err := writeRegistryFile(r.filePath, data); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	// Notify watchers
	r.notifyWatchers(service.Name)

	return nil
}

// GetService retrieves all instances of a specific service.
//
// Parameters:
//   - ctx: The context for the operation
//   - serviceName: The name of the service to retrieve
//
// Returns:
//   - []*kratos_registry.ServiceInstance: A slice of service instances
//   - error: An error if the operation fails
func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*kratos_registry.ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := r.readRegistryFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	instances := data.Services[serviceName]
	result := make([]*kratos_registry.ServiceInstance, 0, len(instances))

	for _, instance := range instances {
		// Convert internal format back to Kratos ServiceInstance
		service := &kratos_registry.ServiceInstance{
			ID:        instance.ID,
			Name:      instance.Name,
			Version:   instance.Version,
			Metadata:  instance.Metadata,
			Endpoints: instance.Endpoints,
		}
		result = append(result, service)
	}

	return result, nil
}

// Watch creates a watcher for service changes.
//
// Parameters:
//   - ctx: The context for the operation
//   - serviceName: The name of the service to watch
//
// Returns:
//   - kratos_registry.Watcher: A watcher for the specified service
//   - error: An error if the watcher cannot be created
func (r *Registry) Watch(ctx context.Context, serviceName string) (kratos_registry.Watcher, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	watcher := NewWatcher(r, serviceName)
	r.watchers[serviceName] = watcher.(*Watcher)

	return watcher, nil
}

// readRegistryFile reads and parses the registry file.
//
// Returns:
//   - *RegistryData: The parsed registry data
//   - error: An error if the file cannot be read or parsed
func (r *Registry) readRegistryFile() (*RegistryData, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	var registryData RegistryData
	if err := json.Unmarshal(data, &registryData); err != nil {
		return nil, err
	}

	return &registryData, nil
}

// writeRegistryFile writes registry data to the file atomically.
//
// Parameters:
//   - filePath: The path to the registry file
//   - data: The registry data to write
//
// Returns:
//   - error: An error if the file cannot be written
func writeRegistryFile(filePath string, data *RegistryData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temporary file first, then rename for atomicity
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return err
	}

	// Atomic rename (works on both Windows and Unix-like systems)
	return os.Rename(tempFile, filePath)
}

// notifyWatchers notifies all watchers about service changes.
// This is a simplified implementation that doesn't use events.
//
// Parameters:
//   - serviceName: The name of the service that changed
func (r *Registry) notifyWatchers(serviceName string) {
	// In this simplified implementation, watchers poll for changes
	// so no explicit notification is needed
}

// removeWatcher removes a watcher from the registry.
//
// Parameters:
//   - serviceName: The name of the service being watched
func (r *Registry) removeWatcher(serviceName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.watchers, serviceName)
}