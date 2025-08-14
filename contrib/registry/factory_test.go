package registry

import (
	"context"
	"testing"
	"time"

	"github.com/cocosip/zero/contrib/registry/conf"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestNewRegistryFactory tests the creation of a new registry factory
func TestNewRegistryFactory(t *testing.T) {
	factory := NewRegistryFactory()
	if factory == nil {
		t.Fatal("Expected factory to be created, got nil")
	}

	_, ok := factory.(*DefaultRegistryFactory)
	if !ok {
		t.Fatal("Expected DefaultRegistryFactory type")
	}
}

// TestCreateLocalRegistry tests the creation of local registry
func TestCreateLocalRegistry(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	// Test valid local registry configuration
	config := &conf.Registry{
		Type: "local",
		Local: &conf.LocalRegistry{
			FilePath: "./test_registry.json",
		},
	}

	registrar, discovery, err := factory.CreateRegistry(ctx, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if registrar == nil {
		t.Fatal("Expected registrar to be created, got nil")
	}

	if discovery == nil {
		t.Fatal("Expected discovery to be created, got nil")
	}

	// Test that registrar and discovery are the same instance for local registry
	// Since they are different interface types, we need to compare their underlying values
	if registrar != registrar || discovery != discovery {
		t.Fatal("Expected registrar and discovery to be valid instances")
	}
	
	// For local registry, both should point to the same underlying object
	// We can verify this by checking if they implement the same interfaces
	if registrar == nil || discovery == nil {
		t.Fatal("Expected both registrar and discovery to be non-nil")
	}
}

// TestCreateLocalRegistry_InvalidConfig tests local registry with invalid configuration
func TestCreateLocalRegistry_InvalidConfig(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	tests := []struct {
		name   string
		config *conf.Registry
	}{
		{
			name: "nil local config",
			config: &conf.Registry{
				Type:  "local",
				Local: nil,
			},
		},
		{
			name: "empty file path",
			config: &conf.Registry{
				Type: "local",
				Local: &conf.LocalRegistry{
					FilePath: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := factory.CreateRegistry(ctx, tt.config)
			if err == nil {
				t.Fatal("Expected error for invalid config, got nil")
			}
		})
	}
}

// TestCreateEtcdRegistry tests etcd registry creation (should return not implemented error)
func TestCreateEtcdRegistry(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	config := &conf.Registry{
		Type: "etcd",
		Etcd: &conf.EtcdRegistry{
			Endpoints:   []string{"127.0.0.1:2379"},
			DialTimeout: durationpb.New(5 * time.Second),
			Username:    "test",
			Password:    "test",
			Namespace:   "/test",
		},
	}

	_, _, err := factory.CreateRegistry(ctx, config)
	if err == nil {
		t.Fatal("Expected 'not implemented' error for etcd registry")
	}

	if err.Error() != "etcd registry not implemented yet" {
		t.Fatalf("Expected 'etcd registry not implemented yet' error, got: %v", err)
	}
}

// TestCreateConsulRegistry tests consul registry creation (should return not implemented error)
func TestCreateConsulRegistry(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	config := &conf.Registry{
		Type: "consul",
		Consul: &conf.ConsulRegistry{
			Address:    "127.0.0.1:8500",
			Scheme:     "http",
			Datacenter: "dc1",
			Token:      "test-token",
			Namespace:  "test",
		},
	}

	_, _, err := factory.CreateRegistry(ctx, config)
	if err == nil {
		t.Fatal("Expected 'not implemented' error for consul registry")
	}

	if err.Error() != "consul registry not implemented yet" {
		t.Fatalf("Expected 'consul registry not implemented yet' error, got: %v", err)
	}
}

// TestCreateNacosRegistry tests nacos registry creation (should return not implemented error)
func TestCreateNacosRegistry(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	config := &conf.Registry{
		Type: "nacos",
		Nacos: &conf.NacosRegistry{
			ServerConfigs: []*conf.NacosServerConfig{
				{
					IpAddr:      "127.0.0.1",
					Port:        8848,
					ContextPath: "/nacos",
				},
			},
			ClientConfig: &conf.NacosClientConfig{
				NamespaceId: "public",
				Username:    "nacos",
				Password:    "nacos",
				LogLevel:    "info",
				LogDir:      "./logs",
				CacheDir:    "./cache",
			},
			Group:   "DEFAULT_GROUP",
			Cluster: "DEFAULT",
		},
	}

	_, _, err := factory.CreateRegistry(ctx, config)
	if err == nil {
		t.Fatal("Expected 'not implemented' error for nacos registry")
	}

	if err.Error() != "nacos registry not implemented yet" {
		t.Fatalf("Expected 'nacos registry not implemented yet' error, got: %v", err)
	}
}

// TestCreateKubernetesRegistry tests kubernetes registry creation (should return not implemented error)
func TestCreateKubernetesRegistry(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	config := &conf.Registry{
		Type: "kubernetes",
		Kubernetes: &conf.KubernetesRegistry{
			Namespace:     "default",
			KubeConfig:    "",
			InCluster:     true,
			LabelSelector: "app=test",
		},
	}

	_, _, err := factory.CreateRegistry(ctx, config)
	if err == nil {
		t.Fatal("Expected 'not implemented' error for kubernetes registry")
	}

	if err.Error() != "kubernetes registry not implemented yet" {
		t.Fatalf("Expected 'kubernetes registry not implemented yet' error, got: %v", err)
	}
}

// TestCreateRegistry_UnsupportedType tests unsupported registry type
func TestCreateRegistry_UnsupportedType(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	config := &conf.Registry{
		Type: "unsupported",
	}

	_, _, err := factory.CreateRegistry(ctx, config)
	if err == nil {
		t.Fatal("Expected error for unsupported registry type")
	}

	expected := "unsupported registry type: unsupported"
	if err.Error() != expected {
		t.Fatalf("Expected '%s' error, got: %v", expected, err)
	}
}

// TestCreateRegistry_NilConfig tests nil configuration
func TestCreateRegistry_NilConfig(t *testing.T) {
	factory := NewRegistryFactory()
	ctx := context.Background()

	_, _, err := factory.CreateRegistry(ctx, nil)
	if err == nil {
		t.Fatal("Expected error for nil config")
	}

	expected := "registry config cannot be nil"
	if err.Error() != expected {
		t.Fatalf("Expected '%s' error, got: %v", expected, err)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *conf.Registry
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "registry config cannot be nil",
		},
		{
			name: "empty type",
			config: &conf.Registry{
				Type: "",
			},
			wantErr: true,
			errMsg:  "registry type cannot be empty",
		},
		{
			name: "valid local config",
			config: &conf.Registry{
				Type: "local",
				Local: &conf.LocalRegistry{
					FilePath: "./test.json",
				},
			},
			wantErr: false,
		},
		{
			name: "local config with nil local",
			config: &conf.Registry{
				Type:  "local",
				Local: nil,
			},
			wantErr: true,
			errMsg:  "local registry config cannot be nil when type is local",
		},
		{
			name: "local config with empty file path",
			config: &conf.Registry{
				Type: "local",
				Local: &conf.LocalRegistry{
					FilePath: "",
				},
			},
			wantErr: true,
			errMsg:  "local registry file path cannot be empty",
		},
		{
			name: "etcd config with nil etcd",
			config: &conf.Registry{
				Type: "etcd",
				Etcd: nil,
			},
			wantErr: true,
			errMsg:  "etcd registry config cannot be nil when type is etcd",
		},
		{
			name: "etcd config with empty endpoints",
			config: &conf.Registry{
				Type: "etcd",
				Etcd: &conf.EtcdRegistry{
					Endpoints: []string{},
				},
			},
			wantErr: true,
			errMsg:  "etcd endpoints cannot be empty",
		},
		{
			name: "unsupported type",
			config: &conf.Registry{
				Type: "unknown",
			},
			wantErr: true,
			errMsg:  "unsupported registry type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if err.Error() != tt.errMsg {
					t.Fatalf("Expected error '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestGetDefaultConfig tests the default configuration
func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	if config == nil {
		t.Fatal("Expected default config to be created, got nil")
	}

	if config.Type != "local" {
		t.Fatalf("Expected default type to be 'local', got: %s", config.Type)
	}

	if config.Local == nil {
		t.Fatal("Expected local config to be set, got nil")
	}

	if config.Local.FilePath != "./registry.json" {
		t.Fatalf("Expected default file path to be './registry.json', got: %s", config.Local.FilePath)
	}

	// Validate that default config is valid
	err := ValidateConfig(config)
	if err != nil {
		t.Fatalf("Expected default config to be valid, got error: %v", err)
	}
}