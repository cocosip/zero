package main

import (
	"context"
	"fmt"
	"log"
	net_http "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	local "github.com/cocosip/zero/contrib/registry/local"
)

// main demonstrates how to use the file-based registry for service registration and discovery.
func main() {
	// Create a temporary directory for the registry file
	tempDir, err := os.MkdirTemp("", "kratos-local-registry")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registryFile := filepath.Join(tempDir, "services.json")
	fmt.Printf("Registry file: %s\n", registryFile)

	// Initialize the local file registry
	reg, err := local.New(registryFile)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}

	// Create HTTP server
	httpSrv := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			recovery.Recovery(),
		),
	)

	// Add a simple HTTP handler
	httpSrv.HandleFunc("/hello", func(w net_http.ResponseWriter, r *net_http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from Kratos with local registry!", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// Create gRPC server
	grpcSrv := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			recovery.Recovery(),
		),
	)

	// Create Kratos app with the local registry
	app := kratos.New(
		kratos.ID("example-service-001"),
		kratos.Name("example.service"),
		kratos.Version("v1.0.0"),
		kratos.Metadata(map[string]string{
			"env":    "development",
			"region": "local",
		}),
		kratos.Server(
			httpSrv,
			grpcSrv,
		),
		kratos.Registrar(reg),
	)

	// Start service discovery example in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second) // Wait for service to register
		demonstateServiceDiscovery(reg)
	}()

	// Start service watcher example in a separate goroutine
	go func() {
		time.Sleep(3 * time.Second) // Wait for service to register
		demonstrateServiceWatcher(reg)
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nReceived shutdown signal, stopping application...")
		if err := app.Stop(); err != nil {
			log.Printf("Failed to stop app: %v", err)
		}
	}()

	// Start the application
	fmt.Println("Starting Kratos application with local file registry...")
	fmt.Println("HTTP server: http://localhost:8000")
	fmt.Println("gRPC server: localhost:9000")
	fmt.Println("Try: curl http://localhost:8000/hello")
	fmt.Println("Press Ctrl+C to stop")

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

// demonstateServiceDiscovery shows how to discover services using the local registry.
//
// Parameters:
//   - reg: The registry instance to use for service discovery
func demonstateServiceDiscovery(reg *local.Registry) {
	ctx := context.Background()

	fmt.Println("\n=== Service Discovery Demo ===")

	// Discover services
	services, err := reg.GetService(ctx, "example.service")
	if err != nil {
		log.Printf("Failed to get services: %v", err)
		return
	}

	fmt.Printf("Found %d instance(s) of 'example.service':\n", len(services))
	for i, service := range services {
		fmt.Printf("  Instance %d:\n", i+1)
		fmt.Printf("    ID: %s\n", service.ID)
		fmt.Printf("    Name: %s\n", service.Name)
		fmt.Printf("    Version: %s\n", service.Version)
		fmt.Printf("    Endpoints: %v\n", service.Endpoints)
		fmt.Printf("    Metadata: %v\n", service.Metadata)
	}
}

// demonstrateServiceWatcher shows how to watch for service changes using the local registry.
//
// Parameters:
//   - reg: The registry instance to use for watching services
func demonstrateServiceWatcher(reg *local.Registry) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n=== Service Watcher Demo ===")
	fmt.Println("Watching for changes to 'example.service' for 30 seconds...")

	// Create a watcher
	watcher, err := reg.Watch(ctx, "example.service")
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}
	defer watcher.Stop()

	// Watch for changes
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Watcher demo completed")
			return
		default:
			// Try to get next event with a timeout
			eventCtx, eventCancel := context.WithTimeout(ctx, 5*time.Second)
			services, err := watcher.Next()
			eventCancel()
			_ = eventCtx // Suppress unused variable warning

			if err != nil {
				if err == context.DeadlineExceeded {
					fmt.Println("No service changes detected in the last 5 seconds")
					continue
				}
				if err == local.ErrWatcherStopped {
					fmt.Println("Watcher stopped")
					return
				}
				log.Printf("Watcher error: %v", err)
				return
			}

			fmt.Printf("Service change detected! Current instances: %d\n", len(services))
			for i, service := range services {
				fmt.Printf("  Instance %d: %s (%v)\n", i+1, service.ID, service.Endpoints)
			}
		}
	}
}