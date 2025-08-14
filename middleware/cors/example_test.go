package cors_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	http2 "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/cocosip/zero/middleware/cors"
)

// ExampleWithConfig demonstrates how to use CORS middleware with proto configuration.
func ExampleWithConfig() {
	// Create CORS configuration using proto message
	config := &cors.CorsConfig{
		AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}

	// Create HTTP server with CORS middleware
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			cors.Server(cors.WithConfig(config)),
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExampleServer_wildcard demonstrates CORS middleware with wildcard origin.
func ExampleServer_wildcard() {
	// Create HTTP server with wildcard CORS
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			cors.Server(
				cors.WithAllowedOrigins("*"),
				cors.WithAllowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS"),
				cors.WithAllowCredentials(false), // Cannot use credentials with wildcard
			),
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-wildcard-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExampleServer_subdomainWildcard demonstrates CORS middleware with subdomain wildcard.
func ExampleServer_subdomainWildcard() {
	// Create HTTP server with subdomain wildcard CORS
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			cors.Server(
				cors.WithAllowedOrigins("*.example.com", "https://localhost:3000"),
				cors.WithAllowedMethods("GET", "POST", "PUT", "DELETE"),
				cors.WithAllowedHeaders("Content-Type", "Authorization"),
				cors.WithAllowCredentials(true),
				cors.WithMaxAge(7200), // 2 hours
			),
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-subdomain-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExampleServer_multipleMiddleware demonstrates combining CORS with other middleware.
func ExampleServer_multipleMiddleware() {
	// Custom logging middleware
	loggingMiddleware := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			fmt.Println("Request received")
			resp, err := handler(ctx, req)
			fmt.Println("Request processed")
			return resp, err
		}
	}

	// Create CORS configuration
	corsConfig := &cors.CorsConfig{
		AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		ExposedHeaders:   []string{"X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}

	// Create HTTP server with multiple middleware
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			loggingMiddleware,                        // Custom logging middleware
			cors.Server(cors.WithConfig(corsConfig)), // CORS middleware
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-multiple-middleware-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExampleServer_development demonstrates CORS configuration for development environment.
func ExampleServer_development() {
	// Development CORS configuration - more permissive
	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			cors.Server(
				cors.WithAllowedOrigins(
					"http://localhost:3000",  // React dev server
					"http://localhost:8080",  // Vue dev server
					"http://localhost:4200",  // Angular dev server
					"http://127.0.0.1:3000",  // Alternative localhost
				),
				cors.WithAllowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"),
				cors.WithAllowedHeaders(
					"Origin",
					"Content-Type",
					"Accept",
					"Authorization",
					"X-Requested-With",
					"X-CSRF-Token",
				),
				cors.WithExposedHeaders("X-Total-Count", "Link"),
				cors.WithAllowCredentials(true),
				cors.WithMaxAge(300), // 5 minutes for development
			),
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-development-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExampleServer_production demonstrates CORS configuration for production environment.
func ExampleServer_production() {
	// Production CORS configuration - more restrictive
	productionConfig := &cors.CorsConfig{
		AllowedOrigins: []string{
			"https://app.mycompany.com",
			"https://admin.mycompany.com",
			"https://mobile.mycompany.com",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"}, // No OPTIONS in production
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}

	httpSrv := http2.NewServer(
		http2.Address(":8080"),
		http2.Middleware(
			cors.Server(cors.WithConfig(productionConfig)),
		),
	)

	// Create Kratos application
	app := kratos.New(
		kratos.Name("cors-production-example"),
		kratos.Server(httpSrv),
	)

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}