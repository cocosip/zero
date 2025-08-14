# Project Rules - Zero (Kratos Extension Library)

## Project Overview

**Project Name**: Zero  
**Repository**: github.com/cocosip/zero  
**Description**: Extension library for Kratos v2.8.4 framework based on github.com/cocosip/utils  
**Language**: Go 1.21  
**Framework**: Kratos v2.8.4  

## Architecture Guidelines

### 1. Project Structure

```
zero/
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── biz/              # Business logic layer
│   ├── data/             # Data access layer
│   ├── service/          # Service layer (gRPC/HTTP handlers)
│   └── conf/             # Configuration structures
├── api/                   # API definitions (protobuf)
├── configs/              # Configuration files
├── pkg/                  # Public library code
├── third_party/          # Third-party proto files
├── test/                 # Integration tests
└── docs/                 # Documentation
```

### 2. Layer Responsibilities

- **Service Layer**: Handle HTTP/gRPC requests, input validation, response formatting
- **Business Layer**: Core business logic, domain rules, use cases
- **Data Layer**: Database operations, external API calls, data persistence

### 3. Dependency Direction

```
Service → Business → Data
```

**Rule**: Upper layers can depend on lower layers, but not vice versa.

## Coding Standards

### 1. Go Code Style

#### File Naming
- Use snake_case for file names: `user_service.go`
- Test files: `user_service_test.go`
- Interface files: `i_user_repository.go`

#### Package Naming
- Use lowercase, single words when possible
- Avoid underscores or mixed caps
- Package name should be descriptive of its purpose

#### Function and Method Naming
- Use PascalCase for exported functions: `CreateUser()`
- Use camelCase for private functions: `validateUserInput()`
- Use descriptive names that explain what the function does

#### Variable Naming
- Use camelCase for variables: `userName`, `userID`
- Use short names for short-lived variables: `i` for loop counters
- Use descriptive names for longer-lived variables

#### Constants
- Use PascalCase for exported constants: `DefaultTimeout`
- Use camelCase for private constants: `defaultRetryCount`
- Group related constants in const blocks

### 2. Comment Standards

#### Package Comments
```go
// Package user provides "user management functionality" for the zero application.
// It includes user creation, authentication, and profile management.
package user
```

#### Function Comments
```go
// CreateUser creates a new user with the provided information.
// It validates the input data, checks for duplicates, and stores the user in the database.
//
// Parameters:
//   - ctx: The context for the operation
//   - req: The user creation request containing "user details"
//
// Returns:
//   - *User: The created user object
//   - error: An error if the operation fails
func CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Implementation
}
```

#### Struct Comments
```go
// User represents a "user entity" in the system.
// It contains all the necessary information for user management.
type User struct {
    ID       int64  `json:"id"`       // Unique identifier for the user
    Username string `json:"username"` // "Username for login"
    Email    string `json:"email"`    // "User email address"
}
```

### 3. Error Handling (Kratos Standard)

#### Proto-based Error Definition
```proto
// api/errors/errors.proto
syntax = "proto3";
package api.errors;

option go_package = "github.com/cocosip/zero/api/errors;errors";

import "errors/errors.proto";

// UserErrorReason defines user-related error reasons
enum UserErrorReason {
  // Set default error code.
  USER_UNSPECIFIED = 0;
  
  // User not found error.
  USER_NOT_FOUND = 1 [(errors.code) = 404];
  
  // Invalid user input error.
  USER_INVALID_INPUT = 2 [(errors.code) = 400];
  
  // User already exists error.
  USER_ALREADY_EXISTS = 3 [(errors.code) = 409];
  
  // User authentication failed error.
  USER_AUTH_FAILED = 4 [(errors.code) = 401];
  
  // User permission denied error.
  USER_PERMISSION_DENIED = 5 [(errors.code) = 403];
}

// CommonErrorReason defines common error reasons
enum CommonErrorReason {
  // Set default error code.
  COMMON_UNSPECIFIED = 0;
  
  // Internal server error.
  INTERNAL_ERROR = 1 [(errors.code) = 500];
  
  // Service unavailable error.
  SERVICE_UNAVAILABLE = 2 [(errors.code) = 503];
  
  // Request timeout error.
  REQUEST_TIMEOUT = 3 [(errors.code) = 408];
  
  // Rate limit exceeded error.
  RATE_LIMIT_EXCEEDED = 4 [(errors.code) = 429];
}
```

#### Generated Error Functions
```go
// Generated from proto file
// api/errors/errors.pb.go (auto-generated)
// Use protoc-gen-go-errors to generate error functions

// IsUserNotFound checks if the error is USER_NOT_FOUND
func IsUserNotFound(err error) bool {
    if err == nil {
        return false
    }
    e := errors.FromError(err)
    return e.Reason == UserErrorReason_USER_NOT_FOUND.String() && e.Code == 404
}

// ErrorUserNotFound creates a USER_NOT_FOUND error
func ErrorUserNotFound(format string, args ...interface{}) *errors.Error {
    return errors.New(404, UserErrorReason_USER_NOT_FOUND.String(), fmt.Sprintf(format, args...))
}

// ErrorUserInvalidInput creates a USER_INVALID_INPUT error
func ErrorUserInvalidInput(format string, args ...interface{}) *errors.Error {
    return errors.New(400, UserErrorReason_USER_INVALID_INPUT.String(), fmt.Sprintf(format, args...))
}
```

#### Error Usage in Service Layer
```go
// Use generated error functions in service layer
func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
    if req.Id <= 0 {
        return nil, ErrorUserInvalidInput("user ID must be positive, got: %d", req.Id)
    }
    
    user, err := s.userRepo.GetByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrorUserNotFound("user with ID %d not found", req.Id)
        }
        return nil, ErrorInternalError("failed to get user: %v", err)
    }
    
    return user, nil
}
```

#### Error Handling in Repository Layer
```go
// Repository layer should return domain errors or wrap system errors
func (r *userRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Return the original error, let service layer handle it
            return nil, err
        }
        // Wrap system errors with context
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    return &user, nil
}
```

#### Error Middleware
```go
// Custom error middleware for additional error handling
func ErrorMiddleware() middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            reply, err := handler(ctx, req)
            if err != nil {
                // Log error details
                log.Context(ctx).Errorw(
                    "msg", "request failed",
                    "error", err.Error(),
                    "request", req,
                )
                
                // Convert unknown errors to internal errors
                if !errors.IsKratosError(err) {
                    return nil, ErrorInternalError("internal server error")
                }
            }
            return reply, err
        }
    }
}
```

#### Error Response Format
```go
// Standard error response format (handled by Kratos automatically)
// HTTP Response:
// {
//   "code": 404,
//   "reason": "USER_NOT_FOUND", 
//   "message": "user with ID 123 not found",
//   "metadata": {}
// }

// gRPC Response:
// status: Code = NotFound
// message: "user with ID 123 not found"
// details: [
//   {
//     "@type": "type.googleapis.com/google.rpc.ErrorInfo",
//     "reason": "USER_NOT_FOUND",
//     "domain": "github.com/cocosip/zero"
//   }
// ]
```

#### Error Generation Commands
```bash
# Generate error code from proto file
protoc --proto_path=. \
       --proto_path=./third_party \
       --go_out=paths=source_relative:. \
       --go-errors_out=paths=source_relative:. \
       api/errors/errors.proto
```

### 4. Context Usage

- Always pass `context.Context` as the first parameter
- Use context for cancellation, timeouts, and request-scoped values
- Don't store contexts in structs

```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Use context for database operations
    return s.userRepo.Create(ctx, req)
}
```

## Testing Standards

### 1. Test File Organization

- Place test files in the same package as the code being tested
- Use `_test.go` suffix for test files
- Group related tests in the same file

### 2. Test Function Naming

```go
// Test function naming pattern: TestFunctionName_Scenario_ExpectedResult
func TestCreateUser_ValidInput_Success(t *testing.T) {}
func TestCreateUser_DuplicateEmail_ReturnsError(t *testing.T) {}
func TestCreateUser_InvalidEmail_ReturnsValidationError(t *testing.T) {}
```

### 3. Test Structure (AAA Pattern)

```go
func TestCreateUser_ValidInput_Success(t *testing.T) {
    // Arrange
    userService := setupUserService(t)
    req := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    // Act
    user, err := userService.CreateUser(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Username, user.Username)
}
```

### 4. Mock Usage

- Use interfaces for dependencies to enable mocking
- Generate mocks using tools like `mockery` or `gomock`
- Place mocks in `mocks/` directory or alongside the interface

### 5. Test Coverage

- Maintain minimum 80% test coverage
- Focus on testing business logic and error paths
- Use table-driven tests for multiple scenarios

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "invalid-email", true},
        {"empty email", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Configuration Management

### 1. Configuration Structure (Kratos Standard)

#### Proto-based Configuration Definition
```proto
// internal/conf/conf.proto
syntax = "proto3";
package kratos.api;

option go_package = "github.com/cocosip/zero/internal/conf;conf";

import "google/protobuf/duration.proto";

// Bootstrap represents the application configuration
message Bootstrap {
  Server server = 1;
  Data data = 2;
}

// Server contains server-related configuration
message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  HTTP http = 1;
  GRPC grpc = 2;
}

// Data contains data layer configuration
message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }
  message Redis {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration read_timeout = 3;
    google.protobuf.Duration write_timeout = 4;
  }
  Database database = 1;
  Redis redis = 2;
}
```

#### Generated Configuration Struct
```go
// Generated from proto file
// internal/conf/conf.pb.go (auto-generated)
type Bootstrap struct {
    Server *Server `protobuf:"bytes,1,opt,name=server,proto3" json:"server,omitempty"`
    Data   *Data   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

type Server struct {
    Http *Server_HTTP `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
    Grpc *Server_GRPC `protobuf:"bytes,2,opt,name=grpc,proto3" json:"grpc,omitempty"`
}
```

### 2. Configuration Loading (Kratos Standard)

#### Basic File-based Configuration
```go
// LoadConfig loads configuration using Kratos config package
// It supports multiple data sources and dynamic configuration updates
func LoadConfig(configPath string) (*conf.Bootstrap, error) {
    c := config.New(
        config.WithSource(
            file.NewSource(configPath),
        ),
    )
    defer c.Close()
    
    if err := c.Load(); err != nil {
        return nil, err
    }
    
    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
        return nil, err
    }
    
    return &bc, nil
}
```

#### Multi-source Configuration
```go
// LoadConfigWithMultipleSources demonstrates loading from multiple sources
// Supports file, environment variables, and remote config centers
func LoadConfigWithMultipleSources() (*conf.Bootstrap, error) {
    c := config.New(
        config.WithSource(
            file.NewSource("configs/config.yaml"),
            env.NewSource("KRATOS_"),
            // Add remote sources like Nacos, Etcd, Consul, etc.
        ),
    )
    defer c.Close()
    
    if err := c.Load(); err != nil {
        return nil, err
    }
    
    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
        return nil, err
    }
    
    return &bc, nil
}
```

#### Dynamic Configuration with Watch
```go
// WatchConfig demonstrates configuration watching for dynamic updates
// Uses atomic operations for safe configuration updates
func WatchConfig(c config.Config) {
    if err := c.Watch("server", func(key string, value config.Value) {
        // Handle configuration changes atomically
        log.Printf("Configuration changed: %s = %v", key, value)
        
        // Reload configuration
        var bc conf.Bootstrap
        if err := c.Scan(&bc); err != nil {
            log.Printf("Failed to scan config: %v", err)
            return
        }
        
        // Apply new configuration atomically
        // Implementation depends on your specific needs
    }); err != nil {
        log.Printf("Failed to watch config: %v", err)
    }
}
```

### 3. Environment-Specific Configuration

#### File Structure
```
configs/
├── config.yaml          # Default configuration
├── config-dev.yaml      # Development environment
├── config-test.yaml     # Testing environment
├── config-prod.yaml     # Production environment
└── config-local.yaml    # Local development (git-ignored)
```

#### Environment-based Loading
```go
// GetConfigPath returns the configuration file path based on environment
// Supports cross-platform file path handling
func GetConfigPath() string {
    env := os.Getenv("KRATOS_ENV")
    if env == "" {
        env = "dev"
    }
    
    // Use filepath.Join for cross-platform compatibility
    configFile := fmt.Sprintf("config-%s.yaml", env)
    return filepath.Join("configs", configFile)
}

// LoadEnvironmentConfig loads configuration based on current environment
func LoadEnvironmentConfig() (*conf.Bootstrap, error) {
    configPath := GetConfigPath()
    return LoadConfig(configPath)
}
```

### 4. Remote Configuration Sources

#### Nacos Integration
```go
// LoadNacosConfig demonstrates Nacos configuration integration
func LoadNacosConfig() (*conf.Bootstrap, error) {
    sc := []constant.ServerConfig{
        *constant.NewServerConfig("127.0.0.1", 8848),
    }
    
    cc := &constant.ClientConfig{
        NamespaceId:         "public",
        TimeoutMs:           5000,
        NotLoadCacheAtStart: true,
        LogDir:              "/tmp/nacos/log",
        CacheDir:            "/tmp/nacos/cache",
        LogLevel:            "info",
    }
    
    client, err := clients.NewConfigClient(
        vo.NacosClientParam{
            ClientConfig:  cc,
            ServerConfigs: sc,
        },
    )
    if err != nil {
        return nil, err
    }
    
    source, err := nacos.NewSource(
        client,
        nacos.WithGroup("DEFAULT_GROUP"),
        nacos.WithDataID("application.yaml"),
    )
    if err != nil {
        return nil, err
    }
    
    c := config.New(config.WithSource(source))
    defer c.Close()
    
    if err := c.Load(); err != nil {
        return nil, err
    }
    
    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
        return nil, err
    }
    
    return &bc, nil
}
```

### 5. Configuration Validation

```go
// ValidateConfig validates the loaded configuration
// Ensures all required fields are properly set
func ValidateConfig(bc *conf.Bootstrap) error {
    if bc.Server == nil {
        return errors.New("server configuration is required")
    }
    
    if bc.Server.Http == nil || bc.Server.Http.Addr == "" {
        return errors.New("HTTP server address is required")
    }
    
    if bc.Server.Grpc == nil || bc.Server.Grpc.Addr == "" {
        return errors.New("gRPC server address is required")
    }
    
    if bc.Data == nil {
        return errors.New("data configuration is required")
    }
    
    return nil
}
```

## Database Guidelines

### 1. Repository Pattern

```go
// IUserRepository defines the interface for user data operations
type IUserRepository interface {
    Create(ctx context.Context, user *User) (*User, error)
    GetByID(ctx context.Context, id int64) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

// userRepository implements IUserRepository
type userRepository struct {
    db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) IUserRepository {
    return &userRepository{db: db}
}
```

### 2. Transaction Management

```go
// Use transactions for operations that modify multiple entities
func (s *UserService) CreateUserWithProfile(ctx context.Context, req *CreateUserRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create user
        user, err := s.userRepo.CreateWithTx(ctx, tx, req.User)
        if err != nil {
            return err
        }
        
        // Create profile
        profile := &Profile{UserID: user.ID, ...}
        return s.profileRepo.CreateWithTx(ctx, tx, profile)
    })
}
```

### 3. Migration Guidelines

- Use descriptive migration file names with timestamps
- Include both up and down migrations
- Test migrations on sample data before applying to production

## API Design Guidelines

### 1. RESTful API Design

- Use HTTP methods appropriately (GET, POST, PUT, DELETE)
- Use plural nouns for resource names: `/api/v1/users`
- Use HTTP status codes correctly
- Include version in URL: `/api/v1/`

### 2. Request/Response Format

```go
// Standard response format
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Error response format
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### 3. Input Validation

```go
// Use validation tags
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// Validate input in service layer
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := s.validator.Struct(req); err != nil {
        return nil, errors.BadRequest("INVALID_INPUT", err.Error())
    }
    // Process request
}
```

## Security Guidelines

### 1. Authentication & Authorization

- Use JWT tokens for stateless authentication
- Implement role-based access control (RBAC)
- Validate permissions at service layer

### 2. Input Sanitization

- Validate all input data
- Use parameterized queries to prevent SQL injection
- Sanitize data before storing or displaying

### 3. Sensitive Data Handling

- Never log sensitive information (passwords, tokens)
- Use environment variables for secrets
- Encrypt sensitive data at rest

## Performance Guidelines

### 1. Database Optimization

- Use database indexes appropriately
- Implement connection pooling
- Use pagination for large result sets
- Avoid N+1 query problems

### 2. Caching Strategy

- Cache frequently accessed data
- Use Redis for distributed caching
- Implement cache invalidation strategies
- Set appropriate TTL values

### 3. Monitoring & Metrics

- Implement health checks
- Monitor key performance indicators
- Use structured logging
- Implement distributed tracing

## Cross-Platform Compatibility

### 1. File Path Handling

```go
// Use filepath.Join for cross-platform path handling
import "path/filepath"

configPath := filepath.Join("configs", "config.yaml")
```

### 2. Build Tags

```go
// Use build tags for platform-specific code
// +build windows

package platform

// Windows-specific implementation
```

### 3. Environment Variables

```go
// Handle different path separators
func getConfigPath() string {
    if runtime.GOOS == "windows" {
        return "configs\\\\config.yaml"
    }
    return "configs/config.yaml"
}
```

## Git Workflow

### 1. Branch Naming

- `feature/feature-name` - New features
- `bugfix/bug-description` - Bug fixes
- `hotfix/critical-fix` - Critical production fixes
- `refactor/component-name` - Code refactoring

### 2. Commit Messages

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(user): add user registration endpoint

Implement user registration with email validation
and password hashing using bcrypt.

Closes #123
```

### 3. Pull Request Guidelines

- Include clear description of changes
- Reference related issues
- Ensure all tests pass
- Request appropriate reviewers
- Update documentation if needed

## Code Review Checklist

### 1. Functionality
- [ ] Code works as intended
- [ ] Edge cases are handled
- [ ] Error handling is appropriate
- [ ] Performance considerations are addressed

### 2. Code Quality
- [ ] Code follows project conventions
- [ ] Functions are reasonably sized
- [ ] Code is readable and well-documented
- [ ] No code duplication

### 3. Testing
- [ ] Adequate test coverage
- [ ] Tests are meaningful and test the right things
- [ ] Tests are maintainable

### 4. Security
- [ ] No sensitive data in code
- [ ] Input validation is implemented
- [ ] Authentication/authorization is correct

## Dependencies Management

### 1. Core Dependencies

- **Kratos v2.8.4**: Main framework
- **github.com/cocosip/utils v0.4.0**: Base utility library
- **GORM**: ORM for database operations
- **Redis**: Caching and session storage

### 2. Dependency Updates

- Review dependencies regularly for security updates
- Test thoroughly before updating major versions
- Document breaking changes in CHANGELOG.md

### 3. Vendor Management

- Use `go mod vendor` for reproducible builds
- Commit vendor directory for critical dependencies
- Keep go.mod and go.sum files clean

---

**Note**: These rules should be followed consistently across the project. Any deviations should be discussed and documented with proper justification.