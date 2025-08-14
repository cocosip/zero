# Local File Registry for Kratos

A local file-based service registry implementation for the [Kratos](https://github.com/go-kratos/kratos) microservice framework. This registry stores service registration information in local JSON files, making it suitable for development environments or deployments where traditional service discovery components (like Consul, etcd, etc.) are not available.

## Features

- **File-based Storage**: Uses local JSON files to store service registration data
- **Cross-platform**: Works on Windows, Linux, and macOS
- **Atomic Operations**: Ensures data consistency with atomic file writes
- **Service Discovery**: Supports service registration, deregistration, and discovery
- **Change Monitoring**: Provides watchers for real-time service change notifications
- **Thread-safe**: Concurrent access protection with mutex locks
- **Zero Dependencies**: No external service discovery infrastructure required

## Installation

```bash
go get github.com/cocosip/zero/contrib/registry/local
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/transport/http"
    "github.com/cocosip/zero/contrib/registry/local"
)

func main() {
    // Create a local file registry
    reg, err := local.New("/tmp/kratos-registry.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create HTTP server
    httpSrv := http.NewServer(http.Address(":8000"))
    
    // Create Kratos app with local registry
    app := kratos.New(
        kratos.ID("my-service-001"),
        kratos.Name("my.service"),
        kratos.Version("v1.0.0"),
        kratos.Server(httpSrv),
        kratos.Registrar(reg), // Use local registry for service registration
    )
    
    // Start the application
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Service Discovery

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/cocosip/zero/contrib/registry/local"
)

func main() {
    // Create registry instance
    reg, err := local.New("/tmp/kratos-registry.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Discover services
    ctx := context.Background()
    services, err := reg.GetService(ctx, "my.service")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d instances of 'my.service'\n", len(services))
    for _, service := range services {
        fmt.Printf("- ID: %s, Endpoints: %v\n", service.ID, service.Endpoints)
    }
}
```

### Watching Service Changes

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/cocosip/zero/contrib/registry/local"
)

func main() {
    reg, err := local.New("/tmp/kratos-registry.json")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Create a watcher for service changes
    watcher, err := reg.Watch(ctx, "my.service")
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Stop()
    
    // Listen for changes
    for {
        services, err := watcher.Next()
        if err != nil {
            log.Printf("Watcher error: %v", err)
            break
        }
        
        fmt.Printf("Service change detected! Current instances: %d\n", len(services))
        for _, service := range services {
            fmt.Printf("- %s: %v\n", service.ID, service.Endpoints)
        }
    }
}
```

## Configuration Options

### Registry Creation

```go
// Create registry with custom file path
reg, err := local.New("/path/to/registry.json")
```

**Parameters:**
- `filePath`: Path to the JSON file where service data will be stored
  - The directory will be created automatically if it doesn't exist
  - On Windows: `C:\\temp\\kratos-registry.json`
  - On Linux/macOS: `/tmp/kratos-registry.json`

## Registry File Format

The registry stores service information in JSON format:

```json
{
  "services": {
    "my.service": [
      {
        "id": "my-service-001",
        "name": "my.service",
        "version": "v1.0.0",
        "metadata": {
          "env": "development",
          "region": "local"
        },
        "endpoints": [
          "http://localhost:8000",
          "grpc://localhost:9000"
        ],
        "timestamp": 1640995200
      }
    ]
  },
  "version": "1.0.0",
  "updated": 1640995200
}
```

## Best Practices

1. **File Location**: Choose a location with appropriate read/write permissions
2. **Backup**: Consider backing up the registry file for production use
3. **Cleanup**: Remove stale service entries periodically
4. **Monitoring**: Monitor file size and implement rotation if needed
5. **Security**: Ensure the registry file is not accessible by unauthorized users

## Limitations and Considerations

- **Single Node**: This registry is designed for single-node or development scenarios
- **No Clustering**: Does not support distributed service discovery across multiple nodes
- **File I/O**: Performance depends on file system I/O capabilities
- **Manual Cleanup**: Stale entries need manual cleanup (no TTL mechanism)
- **No Health Checks**: Does not perform automatic health checking of registered services

## Use Cases

- **Development Environment**: Local development without external dependencies
- **Testing**: Integration and unit testing scenarios
- **Windows Deployment**: Environments where traditional service discovery tools are not available
- **Embedded Systems**: Resource-constrained environments
- **Proof of Concept**: Quick prototyping and demonstrations

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```
   Error: failed to create registry directory: permission denied
   ```
   **Solution**: Ensure the application has write permissions to the specified directory.

2. **File Lock Issues**
   ```
   Error: failed to write registry file: file is locked
   ```
   **Solution**: Ensure no other processes are accessing the registry file.

3. **Invalid JSON**
   ```
   Error: failed to read registry file: invalid character
   ```
   **Solution**: Check if the registry file has been corrupted. Delete and recreate if necessary.

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
import "github.com/go-kratos/kratos/v2/log"

// Enable debug logging
logger := log.With(log.NewStdLogger(os.Stdout),
    "ts", log.DefaultTimestamp,
    "caller", log.DefaultCaller,
)
log.SetLogger(logger)
```

## Example Application

See the [example](./example/main.go) directory for a complete working example that demonstrates:
- Service registration
- Service discovery
- Service watching
- Graceful shutdown

To run the example:

```bash
cd example
go run main.go
```

## License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Changelog

### v1.0.0
- Initial release
- Basic service registration and discovery
- File-based storage with atomic writes
- Service change watching
- Cross-platform support