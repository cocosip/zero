package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/cocosip/zero/contrib/registry"
	registry_conf "github.com/cocosip/zero/contrib/registry/conf"
	"github.com/cocosip/zero/contrib/registry/example/internal/conf"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratos_registry "github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	kratos_http "github.com/go-kratos/kratos/v2/transport/http"
	"gopkg.in/yaml.v3"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "user-service"
	// Version is the version of the compiled software.
	Version string = "v1.0.0"
	// flagconf is the config flag.
	flagconf = flag.String("conf", "configs", "config path, eg: -conf config.yaml")
)

// UserService represents a simple user service implementation for demonstration
type UserService struct {
	log *log.Helper
}

// NewUserService creates a new user service instance
func NewUserService(logger log.Logger) *UserService {
	return &UserService{
		log: log.NewHelper(logger),
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, username, email string) (string, error) {
	s.log.WithContext(ctx).Infof("Creating user: %s (%s)", username, email)
	return fmt.Sprintf("User created successfully: %s (%s)", username, email), nil
}

// GetUser retrieves user information
func (s *UserService) GetUser(ctx context.Context, userID string) (string, error) {
	s.log.WithContext(ctx).Infof("Getting user: %s", userID)
	return fmt.Sprintf("User info for ID: %s - Name: John Doe, Email: john@example.com", userID), nil
}

// ListUsers returns a list of users
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) (string, error) {
	s.log.WithContext(ctx).Infof("Listing users: page=%d, pageSize=%d", page, pageSize)
	return fmt.Sprintf("Users list (page %d, size %d): [User1, User2, User3]", page, pageSize), nil
}

// newApp creates a new Kratos application with the given configuration
func newApp(logger log.Logger, hs *kratos_http.Server, gs *grpc.Server, rr kratos_registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID("user-service-001"),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{
			"env":     "development",
			"region":  "local",
			"zone":    "local",
			"cluster": "default",
		}),
		kratos.Logger(logger),
		kratos.Server(hs, gs),
		kratos.Registrar(rr),
	)
}

// newHTTPServer creates a new HTTP server with the given configuration
func newHTTPServer(c *conf.Server, userSvc *UserService, logger log.Logger) *kratos_http.Server {
	var opts = []kratos_http.ServerOption{
		kratos_http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratos_http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratos_http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratos_http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := kratos_http.NewServer(opts...)

	// Register HTTP routes
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"user-service","version":"` + Version + `","status":"running"}`))
	})

	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	srv.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		switch r.Method {
		case "POST":
			result, err := userSvc.CreateUser(r.Context(), "testuser", "test@example.com")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(result))
		case "GET":
			userID := r.URL.Query().Get("id")
			if userID != "" {
				result, err := userSvc.GetUser(r.Context(), userID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Write([]byte(result))
			} else {
				result, err := userSvc.ListUsers(r.Context(), 1, 10)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Write([]byte(result))
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return srv
}

// newGRPCServer creates a new gRPC server with the given configuration
func newGRPCServer(c *conf.Server, _ *UserService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	// TODO: Register gRPC services here
	// pb.RegisterUserServiceServer(srv, userSvc)

	return srv
}

func main() {
	flag.Parse()

	// Create logger
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", "user-service-001",
		"service.name", Name,
		"service.version", Version,
	)

	// Load configuration
	c := config.New(
		config.WithSource(
			file.NewSource(*flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		log.NewHelper(logger).Fatalf("Failed to load config: %v", err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		log.NewHelper(logger).Fatalf("Failed to scan config: %v", err)
	}

	// Print loaded configuration for debugging
	if configData, err := yaml.Marshal(&bc); err == nil {
		log.NewHelper(logger).Infof("Loaded configuration:\n%s", string(configData))
	}

	// Convert registry configuration
	registryConfig := convertRegistryConfig(bc.Registry)
	
	// Validate registry configuration
	if err := registry.ValidateConfig(registryConfig); err != nil {
		log.NewHelper(logger).Fatalf("Invalid registry config: %v", err)
	}

	// Create registry factory
	factory := registry.NewRegistryFactory()

	// Create registry instances
	ctx := context.Background()
	registrar, _, err := factory.CreateRegistry(ctx, registryConfig)
	if err != nil {
		log.NewHelper(logger).Fatalf("Failed to create registry: %v", err)
	}

	// Print registry information
	log.NewHelper(logger).Infof("Starting application with registry type: %s", bc.Registry.Type)
	switch bc.Registry.Type {
	case "local":
		if bc.Registry.Local != nil {
			log.NewHelper(logger).Infof("Local registry file: %s", bc.Registry.Local.FilePath)
		}
	case "etcd":
		if bc.Registry.Etcd != nil {
			log.NewHelper(logger).Infof("Etcd endpoints: %v", bc.Registry.Etcd.Endpoints)
		}
	case "consul":
		if bc.Registry.Consul != nil {
			log.NewHelper(logger).Infof("Consul address: %s", bc.Registry.Consul.Address)
		}
	case "nacos":
		if bc.Registry.Nacos != nil {
			log.NewHelper(logger).Infof("Nacos servers: %d", len(bc.Registry.Nacos.ServerConfigs))
		}
	case "kubernetes":
		if bc.Registry.Kubernetes != nil {
			log.NewHelper(logger).Infof("Kubernetes namespace: %s", bc.Registry.Kubernetes.Namespace)
		}
	}

	// Create services
	userSvc := NewUserService(logger)

	// Create servers
	httpSrv := newHTTPServer(bc.Server, userSvc, logger)
	grpcSrv := newGRPCServer(bc.Server, userSvc, logger)

	// Create and run application
	app := newApp(logger, httpSrv, grpcSrv, registrar)

	log.NewHelper(logger).Infof("Starting %s version %s", Name, Version)
	log.NewHelper(logger).Infof("HTTP server listening on: %s", bc.Server.Http.Addr)
	log.NewHelper(logger).Infof("gRPC server listening on: %s", bc.Server.Grpc.Addr)

	// Start application
	if err := app.Run(); err != nil {
		log.NewHelper(logger).Fatalf("Failed to run application: %v", err)
	}
}

// convertRegistryConfig converts conf.Registry to registry_conf.Registry
func convertRegistryConfig(src *conf.Registry) *registry_conf.Registry {
	if src == nil {
		return nil
	}
	
	dst := &registry_conf.Registry{
		Type: src.Type,
	}
	
	if src.Local != nil {
		dst.Local = &registry_conf.LocalRegistry{
			FilePath: src.Local.FilePath,
		}
	}
	
	if src.Etcd != nil {
		dst.Etcd = &registry_conf.EtcdRegistry{
			Endpoints: src.Etcd.Endpoints,
			DialTimeout: src.Etcd.DialTimeout,
		}
	}
	
	if src.Consul != nil {
		dst.Consul = &registry_conf.ConsulRegistry{
			Address: src.Consul.Address,
			Scheme:  src.Consul.Scheme,
		}
	}
	
	if src.Nacos != nil {
		dst.Nacos = &registry_conf.NacosRegistry{
			ServerConfigs: make([]*registry_conf.NacosServerConfig, len(src.Nacos.ServerConfigs)),
			ClientConfig:  &registry_conf.NacosClientConfig{},
		}
		
		for i, sc := range src.Nacos.ServerConfigs {
			dst.Nacos.ServerConfigs[i] = &registry_conf.NacosServerConfig{
				IpAddr: sc.IpAddr,
				Port:   sc.Port,
			}
		}
		
		if src.Nacos.ClientConfig != nil {
			dst.Nacos.ClientConfig = &registry_conf.NacosClientConfig{
				NamespaceId: src.Nacos.ClientConfig.NamespaceId,
				Username:    src.Nacos.ClientConfig.Username,
				Password:    src.Nacos.ClientConfig.Password,
				LogLevel:    src.Nacos.ClientConfig.LogLevel,
				LogDir:      src.Nacos.ClientConfig.LogDir,
				CacheDir:    src.Nacos.ClientConfig.CacheDir,
			}
		}
	}
	
	if src.Kubernetes != nil {
		dst.Kubernetes = &registry_conf.KubernetesRegistry{
			KubeConfig: src.Kubernetes.KubeConfig,
			Namespace:  src.Kubernetes.Namespace,
		}
	}
	
	return dst
}