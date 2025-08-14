# CORS Middleware for Kratos

A flexible and configurable CORS (Cross-Origin Resource Sharing) middleware for the Kratos framework. This middleware supports both proto-based configuration and programmatic configuration with wildcard support.

## Features

- ✅ Proto-based configuration support
- ✅ Wildcard origin support (`*`)
- ✅ Subdomain wildcard support (`*.example.com`)
- ✅ Configurable HTTP methods
- ✅ Configurable headers (allowed and exposed)
- ✅ Credentials support
- ✅ Preflight request handling
- ✅ Configurable max age for preflight cache
- ✅ Cross-platform compatibility (Windows, Linux, macOS)

## Installation

```bash
go get github.com/cocosip/zero/middleware/cors
```

## Quick Start

### Basic Usage with Default Configuration

```go
package main

import (
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/cocosip/zero/middleware/cors"
)

func main() {
    // Create HTTP server with default CORS configuration
    httpSrv := http.NewServer(
        http.Address(":8080"),
        http.Middleware(
            cors.Server(), // Default: allows all origins (*)
        ),
    )

    app := kratos.New(
        kratos.Name("my-service"),
        kratos.Server(httpSrv),
    )

    app.Run()
}
```

### Using Configuration File (Recommended)

```go
package main

import (
    "log"
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/cocosip/zero/middleware/cors"
)

func main() {
    // Load configuration from file
    c := config.New(
        config.WithSource(
            file.NewSource("config.yaml"),
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

    httpSrv := http.NewServer(
        http.Address(":8080"),
        http.Middleware(
            corsMiddleware,
        ),
    )

    app := kratos.New(
        kratos.Name("my-service"),
        kratos.Server(httpSrv),
    )

    app.Run()
}
```

**Configuration file example (config.yaml):**

```yaml
middleware:
  cors:
    allowed_origins:
      - "https://example.com"
      - "https://app.example.com"
      - "https://*.example.com"  # Subdomain wildcard
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allowed_headers:
      - "Content-Type"
      - "Authorization"
      - "X-Requested-With"
    exposed_headers:
      - "X-Total-Count"
    allow_credentials: true
    max_age: 3600  # 1 hour
```

### Using Proto Configuration

```go
package main

import (
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/cocosip/zero/middleware/cors"
)

func main() {
    // Create CORS configuration using proto message
    config := &cors.CorsConfig{
        AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        ExposedHeaders:   []string{"X-Total-Count"},
        AllowCredentials: true,
        MaxAge:           3600, // 1 hour
    }

    httpSrv := http.NewServer(
        http.Address(":8080"),
        http.Middleware(
            cors.Server(cors.WithConfig(config)),
        ),
    )

    app := kratos.New(
        kratos.Name("my-service"),
        kratos.Server(httpSrv),
    )

    app.Run()
}
```

### Using Programmatic Configuration

```go
package main

import (
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/cocosip/zero/middleware/cors"
)

func main() {
    httpSrv := http.NewServer(
        http.Address(":8080"),
        http.Middleware(
            cors.Server(
                cors.WithAllowedOrigins("https://example.com", "https://app.example.com"),
                cors.WithAllowedMethods("GET", "POST", "PUT", "DELETE"),
                cors.WithAllowedHeaders("Content-Type", "Authorization"),
                cors.WithExposedHeaders("X-Total-Count"),
                cors.WithAllowCredentials(true),
                cors.WithMaxAge(3600),
            ),
        ),
    )

    app := kratos.New(
        kratos.Name("my-service"),
        kratos.Server(httpSrv),
    )

    app.Run()
}
```

## Configuration Options

### Proto Configuration

The `CorsConfig` proto message supports the following fields:

```proto
message CorsConfig {
  repeated string allowed_origins = 1;    // Allowed origins (use "*" for all)
  repeated string allowed_methods = 2;    // Allowed HTTP methods
  repeated string allowed_headers = 3;    // Allowed request headers
  repeated string exposed_headers = 4;    // Headers exposed to client
  bool allow_credentials = 5;             // Allow credentials
  int32 max_age = 6;                     // Preflight cache max age (seconds)
}
```

### Programmatic Configuration Functions

| Function | Description | Example |
|----------|-------------|----------|
| `WithConfig(config *CorsConfig)` | Use proto configuration | `cors.WithConfig(config)` |
| `WithAllowedOrigins(origins ...string)` | Set allowed origins | `cors.WithAllowedOrigins("*")` |
| `WithAllowedMethods(methods ...string)` | Set allowed HTTP methods | `cors.WithAllowedMethods("GET", "POST")` |
| `WithAllowedHeaders(headers ...string)` | Set allowed request headers | `cors.WithAllowedHeaders("Content-Type")` |
| `WithExposedHeaders(headers ...string)` | Set exposed response headers | `cors.WithExposedHeaders("X-Total-Count")` |
| `WithAllowCredentials(allow bool)` | Enable/disable credentials | `cors.WithAllowCredentials(true)` |
| `WithMaxAge(maxAge int32)` | Set preflight cache duration | `cors.WithMaxAge(3600)` |

## Origin Patterns

### Exact Match
```go
cors.WithAllowedOrigins("https://example.com", "https://app.example.com")
```

### Wildcard (All Origins)
```go
cors.WithAllowedOrigins("*")
```
**Note**: When using `*`, credentials cannot be enabled for security reasons.

### Subdomain Wildcard
```go
cors.WithAllowedOrigins("*.example.com")
```
This allows:
- `https://api.example.com`
- `https://app.example.com`
- `https://admin.example.com`
- `https://example.com` (exact domain)

## Environment-Specific Configurations

### Development Environment

```go
// More permissive for development
cors.Server(
    cors.WithAllowedOrigins(
        "http://localhost:3000",  // React
        "http://localhost:8080",  // Vue
        "http://localhost:4200",  // Angular
    ),
    cors.WithAllowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"),
    cors.WithAllowedHeaders("*"),
    cors.WithAllowCredentials(true),
    cors.WithMaxAge(300), // 5 minutes
)
```

### Production Environment

```go
// More restrictive for production
config := &cors.CorsConfig{
    AllowedOrigins: []string{
        "https://app.mycompany.com",
        "https://admin.mycompany.com",
    },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    ExposedHeaders:   []string{"X-RateLimit-Remaining"},
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
cors.Server(cors.WithConfig(config))
```

## Default Configuration

When no configuration is provided, the middleware uses these defaults:

```go
{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
    AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
    ExposedHeaders:   []string{},
    AllowCredentials: false,
    MaxAge:           86400, // 24 hours
}
```

## Preflight Requests

The middleware automatically handles CORS preflight requests (OPTIONS method) by:

1. Setting appropriate CORS headers
2. Returning `204 No Content` status
3. Preventing the request from reaching your handlers

## Security Considerations

1. **Wildcard Origins**: Using `*` for origins disables credential support for security reasons
2. **Credentials**: Only enable credentials when necessary and with specific origins
3. **Headers**: Be specific about allowed headers in production
4. **Methods**: Only allow the HTTP methods your API actually supports

## Testing

Run the tests:

```bash
go test ./middleware/cors/...
```

Run tests with coverage:

```bash
go test -cover ./middleware/cors/...
```

## Examples

See the `example_test.go` file for comprehensive usage examples including:

- Basic configuration
- Wildcard origins
- Subdomain wildcards
- Multiple middleware integration
- Environment-specific configurations

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass
2. Code follows Go conventions
3. Comments are in English with Chinese terms in quotes
4. Cross-platform compatibility is maintained

## License

This project is part of the Zero library for Kratos framework.