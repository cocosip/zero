package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	http2 "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/cocosip/zero/middleware/cors"
)

// Bootstrap represents the application configuration structure
type Bootstrap struct {
	Server     *Server     `yaml:"server"`
	Middleware *Middleware `yaml:"middleware"`
}

// Server contains server-related configuration
type Server struct {
	HTTP *ServerHTTP `yaml:"http"`
	GRPC *ServerGRPC `yaml:"grpc"`
}

// ServerHTTP contains HTTP server configuration
type ServerHTTP struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

// ServerGRPC contains gRPC server configuration
type ServerGRPC struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

// Middleware contains middleware configuration
type Middleware struct {
	Cors *cors.CorsConfig `yaml:"cors"`
}

// ExampleServerWithConfigFromFile demonstrates how to use CORS middleware with configuration loaded from file
func ExampleServerWithConfigFromFile() {
	// Load configuration from file
	c := config.New(
		config.WithSource(
			file.NewSource("config_example.yaml"),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create CORS middleware from configuration
	corsMiddleware, err := cors.ServerWithConfig(c, "middleware.cors")
	if err != nil {
		log.Fatalf("Failed to create CORS middleware: %v", err)
	}

	// Create HTTP server with CORS middleware
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			corsMiddleware,
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-config-example"),
		kratos.Server(httpSrv),
	)

	// Handle graceful shutdown
	c_signal := make(chan os.Signal, 1)
	signal.Notify(c_signal, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c_signal
		log.Println("Received shutdown signal, stopping application...")
		if err := app.Stop(); err != nil {
			log.Printf("Failed to stop app: %v", err)
		}
	}()

	// Start the application
	log.Println("Starting application with CORS configuration from file...")
	log.Println("HTTP server: http://localhost:8080")
	log.Println("Press Ctrl+C to stop")

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

// ExampleServerWithConfigDynamic demonstrates how to use CORS middleware with dynamic configuration
func ExampleServerWithConfigDynamic() {
	// Create in-memory configuration
	c := config.New()
	defer c.Close()

	// Set CORS configuration programmatically
	corsConfig := &cors.CorsConfig{
		AllowedOrigins: []string{
			"https://example.com",
			"https://app.example.com",
			"http://localhost:3000",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "Authorization",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	// Create CORS middleware directly with config
	corsMiddleware := cors.Server(cors.WithConfig(corsConfig))

	// Create HTTP server with CORS middleware
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			corsMiddleware,
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-dynamic-config-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	log.Println("Starting application with dynamic CORS configuration...")
	log.Println("HTTP server: http://localhost:8080")

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

// ExampleServerWithEnvironmentConfig demonstrates environment-specific CORS configuration
func ExampleServerWithEnvironmentConfig() {
	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load environment-specific configuration
	configFile := "config_" + env + ".yaml"
	c := config.New(
		config.WithSource(
			file.NewSource(configFile),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Create CORS middleware from environment-specific configuration
	corsMiddleware, err := cors.ServerWithConfig(c, "middleware.cors")
	if err != nil {
		log.Fatalf("Failed to create CORS middleware: %v", err)
	}

	// Create HTTP server with CORS middleware
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			corsMiddleware,
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-env-config-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	log.Printf("Starting application with %s CORS configuration...", env)
	log.Println("HTTP server: http://localhost:8080")
	log.Printf("Configuration file: %s", configFile)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

func main() {
	// Run the file-based configuration example
	ExampleServerWithConfigFromFile()
}
