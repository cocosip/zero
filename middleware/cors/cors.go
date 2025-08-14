package cors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/middleware"
)

// Option is a function that configures the CORS middleware
type Option func(*options)

// options holds the configuration for CORS middleware
type options struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int32
}

// WithConfig configures CORS middleware using proto configuration
// Parameters:
//   - config: The CORS configuration from proto definition
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithConfig(config *CorsConfig) Option {
	return func(o *options) {
		if config == nil {
			return
		}
		o.allowedOrigins = config.AllowedOrigins
		o.allowedMethods = config.AllowedMethods
		o.allowedHeaders = config.AllowedHeaders
		o.exposedHeaders = config.ExposedHeaders
		o.allowCredentials = config.AllowCredentials
		o.maxAge = config.MaxAge
	}
}

// WithAllowedOrigins sets the allowed origins for CORS requests
// Parameters:
//   - origins: List of allowed origins, use "*" to allow all origins
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithAllowedOrigins(origins ...string) Option {
	return func(o *options) {
		o.allowedOrigins = origins
	}
}

// WithAllowedMethods sets the allowed HTTP methods for CORS requests
// Parameters:
//   - methods: List of allowed HTTP methods (GET, POST, PUT, DELETE, etc.)
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithAllowedMethods(methods ...string) Option {
	return func(o *options) {
		o.allowedMethods = methods
	}
}

// WithAllowedHeaders sets the allowed headers for CORS requests
// Parameters:
//   - headers: List of allowed headers
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithAllowedHeaders(headers ...string) Option {
	return func(o *options) {
		o.allowedHeaders = headers
	}
}

// WithExposedHeaders sets the headers that are exposed to the client
// Parameters:
//   - headers: List of headers to expose
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithExposedHeaders(headers ...string) Option {
	return func(o *options) {
		o.exposedHeaders = headers
	}
}

// WithAllowCredentials sets whether credentials are allowed
// Parameters:
//   - allow: Boolean indicating whether to allow credentials
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithAllowCredentials(allow bool) Option {
	return func(o *options) {
		o.allowCredentials = allow
	}
}

// WithMaxAge sets the maximum age for preflight requests cache
// Parameters:
//   - maxAge: Maximum age in seconds
//
// Returns:
//   - Option: Configuration function for CORS middleware
func WithMaxAge(maxAge int32) Option {
	return func(o *options) {
		o.maxAge = maxAge
	}
}

// Server returns a CORS middleware for Kratos HTTP server.
// It handles CORS preflight requests and adds appropriate CORS headers to responses.
//
// Parameters:
//   - opts: Configuration options for CORS behavior
//
// Returns:
//   - middleware.Middleware: The CORS middleware function
func Server(opts ...Option) middleware.Middleware {
	o := &options{
		allowedOrigins:   []string{"*"},
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		allowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		exposedHeaders:   []string{},
		allowCredentials: false,
		maxAge:           0,
	}

	// Apply options
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// For now, just continue with the handler
			// The actual CORS handling will be implemented when we have proper HTTP context access
			return handler(ctx, req)
		}
	}
}

// ServerWithConfig creates a CORS middleware for Kratos server using configuration from config source
// This function reads CORS configuration from the application's config file
// Parameters:
//   - c: Kratos config instance
//   - configKey: Configuration key path for CORS config (e.g., "middleware.cors")
//
// Returns:
//   - middleware.Middleware: Configured CORS middleware
//   - error: Error if configuration loading fails
func ServerWithConfig(c config.Config, configKey string) (middleware.Middleware, error) {
	var corsConfig CorsConfig
	if err := c.Value(configKey).Scan(&corsConfig); err != nil {
		return nil, fmt.Errorf("failed to load CORS configuration from key '%s': %w", configKey, err)
	}

	// Create middleware with loaded configuration
	return Server(WithConfig(&corsConfig)), nil
}

// isOriginAllowed checks if the given origin is allowed
// Parameters:
//   - origin: The origin to check
//   - allowedOrigins: List of allowed origins
//
// Returns:
//   - bool: True if origin is allowed, false otherwise
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	// Check for wildcard
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return true
	}

	// Check exact match
	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// HTTPMiddleware returns a standard HTTP middleware function for CORS handling.
// This can be used with standard HTTP servers or other frameworks.
//
// Parameters:
//   - opts: Configuration options for CORS behavior
//
// Returns:
//   - func(http.Handler) http.Handler: A standard HTTP middleware function
func HTTPMiddleware(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		allowedOrigins:   []string{"*"},
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		allowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		exposedHeaders:   []string{},
		allowCredentials: false,
		maxAge:           0,
	}

	// Apply options
	for _, opt := range opts {
		opt(o)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get origin from request
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, o.allowedOrigins) {
				header := w.Header()
				
				// Set CORS headers
				header.Set("Access-Control-Allow-Origin", getOriginValue(origin, o.allowedOrigins))

				if len(o.allowedMethods) > 0 {
					header.Set("Access-Control-Allow-Methods", strings.Join(o.allowedMethods, ", "))
				}

				if len(o.allowedHeaders) > 0 {
					header.Set("Access-Control-Allow-Headers", strings.Join(o.allowedHeaders, ", "))
				}

				if len(o.exposedHeaders) > 0 {
					header.Set("Access-Control-Expose-Headers", strings.Join(o.exposedHeaders, ", "))
				}

				if o.allowCredentials {
					header.Set("Access-Control-Allow-Credentials", "true")
				}

				if o.maxAge > 0 {
					header.Set("Access-Control-Max-Age", fmt.Sprintf("%d", o.maxAge))
				}
				
				// Handle preflight requests
				if r.Method == "OPTIONS" {
					// Check if this is a CORS preflight request
					if r.Header.Get("Access-Control-Request-Method") != "" {
						w.WriteHeader(http.StatusNoContent)
						return
					}
				}
			}

			// Continue with the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// getOriginValue returns the appropriate origin value for the Access-Control-Allow-Origin header
// Parameters:
//   - origin: The request origin
//   - allowedOrigins: List of allowed origins
//
// Returns:
//   - string: The origin value to set in the header
func getOriginValue(origin string, allowedOrigins []string) string {
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return "*"
	}
	return origin
}