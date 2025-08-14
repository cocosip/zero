package registry

import (
	"context"
	"fmt"

	"github.com/cocosip/zero/contrib/registry/conf"
	"github.com/cocosip/zero/contrib/registry/local"
	kratos_registry "github.com/go-kratos/kratos/v2/registry"
)

// RegistryFactory defines the interface for creating registry instances
// It provides a unified way to create different types of registries based on configuration
type RegistryFactory interface {
	// CreateRegistry creates a registry instance based on the provided configuration
	// Parameters:
	//   - ctx: The context for the operation
	//   - config: The registry configuration containing type and specific settings
	// Returns:
	//   - kratos_registry.Registrar: The registrar instance for service registration
	//   - kratos_registry.Discovery: The discovery instance for service discovery
	//   - error: An error if the creation fails
	CreateRegistry(ctx context.Context, config *conf.Registry) (kratos_registry.Registrar, kratos_registry.Discovery, error)
}

// DefaultRegistryFactory is the default implementation of RegistryFactory
// It supports creating local, etcd, consul, nacos, and kubernetes registries
type DefaultRegistryFactory struct{}

// NewRegistryFactory creates a new instance of DefaultRegistryFactory
// Returns:
//   - RegistryFactory: A new registry factory instance
func NewRegistryFactory() RegistryFactory {
	return &DefaultRegistryFactory{}
}

// CreateRegistry creates a registry instance based on the provided configuration
// It validates the configuration and creates the appropriate registry type
// Parameters:
//   - ctx: The context for the operation
//   - config: The registry configuration containing type and specific settings
//
// Returns:
//   - kratos_registry.Registrar: The registrar instance for service registration
//   - kratos_registry.Discovery: The discovery instance for service discovery
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) CreateRegistry(ctx context.Context, config *conf.Registry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("registry config cannot be nil")
	}

	switch config.Type {
	case "local":
		return f.createLocalRegistry(ctx, config.Local)
	case "etcd":
		return f.createEtcdRegistry(ctx, config.Etcd)
	case "consul":
		return f.createConsulRegistry(ctx, config.Consul)
	case "nacos":
		return f.createNacosRegistry(ctx, config.Nacos)
	case "kubernetes":
		return f.createKubernetesRegistry(ctx, config.Kubernetes)
	default:
		return nil, nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}

// createLocalRegistry creates a local file-based registry instance
// Parameters:
//   - ctx: The context for the operation
//   - config: The local registry configuration
//
// Returns:
//   - kratos_registry.Registrar: The local registrar instance
//   - kratos_registry.Discovery: The local discovery instance
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) createLocalRegistry(_ context.Context, config *conf.LocalRegistry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("local registry config cannot be nil")
	}

	if config.FilePath == "" {
		return nil, nil, fmt.Errorf("local registry file path cannot be empty")
	}

	// Create local registry instance
	registry, err := local.New(config.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local registry: %w", err)
	}
	return registry, registry, nil
}

// createEtcdRegistry creates an etcd-based registry instance
// Parameters:
//   - ctx: The context for the operation
//   - config: The etcd registry configuration
//
// Returns:
//   - kratos_registry.Registrar: The etcd registrar instance
//   - kratos_registry.Discovery: The etcd discovery instance
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) createEtcdRegistry(_ context.Context, config *conf.EtcdRegistry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("etcd registry config cannot be nil")
	}

	if len(config.Endpoints) == 0 {
		return nil, nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	// TODO: Implement etcd registry creation
	// This would require importing etcd client library and creating etcd registry
	return nil, nil, fmt.Errorf("etcd registry not implemented yet")
}

// createConsulRegistry creates a consul-based registry instance
// Parameters:
//   - ctx: The context for the operation
//   - config: The consul registry configuration
//
// Returns:
//   - kratos_registry.Registrar: The consul registrar instance
//   - kratos_registry.Discovery: The consul discovery instance
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) createConsulRegistry(_ context.Context, config *conf.ConsulRegistry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("consul registry config cannot be nil")
	}

	if config.Address == "" {
		return nil, nil, fmt.Errorf("consul address cannot be empty")
	}

	// TODO: Implement consul registry creation
	// This would require importing consul client library and creating consul registry
	return nil, nil, fmt.Errorf("consul registry not implemented yet")
}

// createNacosRegistry creates a nacos-based registry instance
// Parameters:
//   - ctx: The context for the operation
//   - config: The nacos registry configuration
//
// Returns:
//   - kratos_registry.Registrar: The nacos registrar instance
//   - kratos_registry.Discovery: The nacos discovery instance
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) createNacosRegistry(_ context.Context, config *conf.NacosRegistry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("nacos registry config cannot be nil")
	}

	if len(config.ServerConfigs) == 0 {
		return nil, nil, fmt.Errorf("nacos server configs cannot be empty")
	}

	// TODO: Implement nacos registry creation
	// This would require importing nacos client library and creating nacos registry
	return nil, nil, fmt.Errorf("nacos registry not implemented yet")
}

// createKubernetesRegistry creates a kubernetes-based registry instance
// Parameters:
//   - ctx: The context for the operation
//   - config: The kubernetes registry configuration
//
// Returns:
//   - kratos_registry.Registrar: The kubernetes registrar instance
//   - kratos_registry.Discovery: The kubernetes discovery instance
//   - error: An error if the creation fails
func (f *DefaultRegistryFactory) createKubernetesRegistry(_ context.Context, config *conf.KubernetesRegistry) (kratos_registry.Registrar, kratos_registry.Discovery, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("kubernetes registry config cannot be nil")
	}

	// TODO: Implement kubernetes registry creation
	// This would require importing kubernetes client library and creating kubernetes registry
	return nil, nil, fmt.Errorf("kubernetes registry not implemented yet")
}

// ValidateConfig validates the registry configuration
// Parameters:
//   - config: The registry configuration to validate
//
// Returns:
//   - error: An error if the validation fails
func ValidateConfig(config *conf.Registry) error {
	if config == nil {
		return fmt.Errorf("registry config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("registry type cannot be empty")
	}

	switch config.Type {
	case "local":
		if config.Local == nil {
			return fmt.Errorf("local registry config cannot be nil when type is local")
		}
		if config.Local.FilePath == "" {
			return fmt.Errorf("local registry file path cannot be empty")
		}
	case "etcd":
		if config.Etcd == nil {
			return fmt.Errorf("etcd registry config cannot be nil when type is etcd")
		}
		if len(config.Etcd.Endpoints) == 0 {
			return fmt.Errorf("etcd endpoints cannot be empty")
		}
	case "consul":
		if config.Consul == nil {
			return fmt.Errorf("consul registry config cannot be nil when type is consul")
		}
		if config.Consul.Address == "" {
			return fmt.Errorf("consul address cannot be empty")
		}
	case "nacos":
		if config.Nacos == nil {
			return fmt.Errorf("nacos registry config cannot be nil when type is nacos")
		}
		if len(config.Nacos.ServerConfigs) == 0 {
			return fmt.Errorf("nacos server configs cannot be empty")
		}
	case "kubernetes":
		if config.Kubernetes == nil {
			return fmt.Errorf("kubernetes registry config cannot be nil when type is kubernetes")
		}
	default:
		return fmt.Errorf("unsupported registry type: %s", config.Type)
	}

	return nil
}

// GetDefaultConfig returns a default registry configuration for local registry
// Returns:
//   - *conf.Registry: A default registry configuration
func GetDefaultConfig() *conf.Registry {
	return &conf.Registry{
		Type: "local",
		Local: &conf.LocalRegistry{
			FilePath: "./registry.json",
		},
	}
}
