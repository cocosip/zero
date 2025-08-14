package cors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsOriginAllowed_ExactMatch tests exact origin matching
func TestIsOriginAllowed_ExactMatch(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "exact match allowed",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			expected:       true,
		},
		{
			name:           "exact match not allowed",
			origin:         "https://notallowed.com",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			expected:       false,
		},
		{
			name:           "empty allowed origins",
			origin:         "https://example.com",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsOriginAllowed_Wildcard tests wildcard origin matching
func TestIsOriginAllowed_Wildcard(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "wildcard allows all",
			origin:         "https://any-domain.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "wildcard with empty origin",
			origin:         "",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsOriginAllowed_WildcardSubdomain tests wildcard subdomain matching
func TestIsOriginAllowed_WildcardSubdomain(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "wildcard subdomain match",
			origin:         "https://api.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard subdomain exact domain match",
			origin:         "https://example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard subdomain no match",
			origin:         "https://test.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       false,
		},
		{
			name:           "multiple level subdomain match",
			origin:         "https://api.v1.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetOriginValue tests getting the correct origin value for headers
func TestGetOriginValue(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       string
	}{
		{
			name:           "wildcard returns asterisk",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       "*",
		},
		{
			name:           "specific origin returns origin",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			expected:       "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOriginValue(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestWithConfig tests the WithConfig option function
func TestWithConfig(t *testing.T) {
	// Arrange
	config := &CorsConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	opts := &options{}
	optionFunc := WithConfig(config)

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, []string{"https://example.com"}, opts.allowedOrigins)
	assert.Equal(t, []string{"GET", "POST"}, opts.allowedMethods)
	assert.Equal(t, []string{"Content-Type"}, opts.allowedHeaders)
	assert.Equal(t, []string{"X-Total-Count"}, opts.exposedHeaders)
	assert.True(t, opts.allowCredentials)
	assert.Equal(t, int32(3600), opts.maxAge)
}

// TestWithConfig_NilConfig tests WithConfig with nil configuration
func TestWithConfig_NilConfig(t *testing.T) {
	// Arrange
	opts := &options{
		allowedOrigins: []string{"original"},
	}
	optionFunc := WithConfig(nil)

	// Act
	optionFunc(opts)

	// Assert - should not change original values
	assert.Equal(t, []string{"original"}, opts.allowedOrigins)
}

// TestWithAllowedOrigins tests the WithAllowedOrigins option function
func TestWithAllowedOrigins(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithAllowedOrigins("https://example.com", "https://test.com")

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, []string{"https://example.com", "https://test.com"}, opts.allowedOrigins)
}

// TestWithAllowedMethods tests the WithAllowedMethods option function
func TestWithAllowedMethods(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithAllowedMethods("GET", "POST", "PUT")

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, []string{"GET", "POST", "PUT"}, opts.allowedMethods)
}

// TestWithAllowedHeaders tests the WithAllowedHeaders option function
func TestWithAllowedHeaders(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithAllowedHeaders("Content-Type", "Authorization")

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, []string{"Content-Type", "Authorization"}, opts.allowedHeaders)
}

// TestWithExposedHeaders tests the WithExposedHeaders option function
func TestWithExposedHeaders(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithExposedHeaders("X-Total-Count", "X-Page-Count")

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, []string{"X-Total-Count", "X-Page-Count"}, opts.exposedHeaders)
}

// TestWithAllowCredentials tests the WithAllowCredentials option function
func TestWithAllowCredentials(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithAllowCredentials(true)

	// Act
	optionFunc(opts)

	// Assert
	assert.True(t, opts.allowCredentials)
}

// TestWithMaxAge tests the WithMaxAge option function
func TestWithMaxAge(t *testing.T) {
	// Arrange
	opts := &options{}
	optionFunc := WithMaxAge(7200)

	// Act
	optionFunc(opts)

	// Assert
	assert.Equal(t, int32(7200), opts.maxAge)
}

// TestServer_CreatesMiddleware tests that Server function creates middleware
func TestServer_CreatesMiddleware(t *testing.T) {
	// Act
	middleware := Server()

	// Assert
	assert.NotNil(t, middleware)
}

// TestServer_WithOptions tests Server function with various options
func TestServer_WithOptions(t *testing.T) {
	// Act
	middleware := Server(
		WithAllowedOrigins("https://example.com"),
		WithAllowedMethods("GET", "POST"),
		WithAllowCredentials(true),
	)

	// Assert
	assert.NotNil(t, middleware)
}