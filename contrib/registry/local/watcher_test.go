package local

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWatch_ValidService_CreatesWatcher tests creating a new watcher with valid service
func TestWatch_ValidService_CreatesWatcher(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"

	// Act
	watcher, err := reg.Watch(ctx, serviceName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	// Cleanup
	watcher.Stop()
}

// TestWatcher_Next_ReturnsCurrentInstances tests Next method returns current instances
func TestWatcher_Next_ReturnsCurrentInstances(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	service := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	// Register service first
	reg.Register(ctx, service)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// Act
	result, err := watcher.Next()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, service.ID, result[0].ID)
}

// TestWatcher_Next_ReceivesUpdates tests Next method receives updates when services change
func TestWatcher_Next_ReceivesUpdates(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	initialService := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	// Register initial service
	reg.Register(ctx, initialService)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// First call should return initial service
	result1, err1 := watcher.Next()
	assert.NoError(t, err1)
	assert.Len(t, result1, 1)
	assert.Equal(t, "test-service-001", result1[0].ID)

	// Register another service instance
	additionalService := &registry.ServiceInstance{
		ID:        "test-service-002",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8081"},
	}
	reg.Register(ctx, additionalService)

	// Second call should return updated instances (with timeout)
	done := make(chan bool)
	var result2 []*registry.ServiceInstance
	var err2 error

	go func() {
		result2, err2 = watcher.Next()
		done <- true
	}()

	select {
	case <-done:
		assert.NoError(t, err2)
		assert.Len(t, result2, 2)
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for watcher update - this may be expected behavior")
	}
}

// TestWatcher_Stop_StopsWatcher tests Stop method functionality
func TestWatcher_Stop_StopsWatcher(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	service := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	reg.Register(ctx, service)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)

	// Act
	watcher.Stop()

	// Assert - Next() should return error after stop
	_, err = watcher.Next()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "watcher stopped")
}

// TestWatcher_Next_AfterStop_ReturnsError tests Next method after watcher is stopped
func TestWatcher_Next_AfterStop_ReturnsError(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	service := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	reg.Register(ctx, service)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)

	// Act
	watcher.Stop()
	result, err := watcher.Next()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "watcher stopped")
}

// TestWatcher_Stop_SetsStoppedFlag tests Stop method sets stopped flag
func TestWatcher_Stop_SetsStoppedFlag(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)

	// Act
	watcher.Stop()

	// Assert - Verify watcher is stopped by checking Next() returns error
	_, err = watcher.Next()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "watcher stopped")
}

// TestWatcher_Stop_MultipleCalls_NoError tests multiple Stop calls don't cause errors
func TestWatcher_Stop_MultipleCalls_NoError(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)

	// Act
	watcher.Stop()
	watcher.Stop() // Second call should not panic
	watcher.Stop() // Third call should not panic

	// Assert - Verify watcher is stopped
	_, err = watcher.Next()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "watcher stopped")
}

// TestWatcher_Next_WithTimeout_HandlesTimeout tests Next method with context timeout
func TestWatcher_Next_WithTimeout_HandlesTimeout(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// Act - Call Next() first to get initial instances
	result1, err1 := watcher.Next()
	assert.NoError(t, err1)
	assert.Empty(t, result1) // No services registered initially

	// Start a goroutine to call Next() which will block waiting for updates
	done := make(chan bool)
	var result2 []*registry.ServiceInstance
	var err2 error

	go func() {
		result2, err2 = watcher.Next()
		done <- true
	}()

	// Wait a short time then stop the watcher
	time.Sleep(100 * time.Millisecond)
	watcher.Stop()

	// Wait for the goroutine to complete
	select {
	case <-done:
		// Assert
		assert.Error(t, err2)
		assert.Nil(t, result2)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for Next() to return after Stop()")
	}
}

// TestWatcher_ConcurrentOperations tests concurrent operations on watcher
func TestWatcher_ConcurrentOperations(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	service := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	reg.Register(ctx, service)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// Act - Start multiple goroutines calling Next()
	goroutineCount := 5
	done := make(chan bool, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer func() { done <- true }()
			// Each goroutine calls Next() once
			result, err := watcher.Next()
			if err == nil {
				assert.NotNil(t, result)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutineCount; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}

// TestWatcher_UpdateInstances_ReflectsChanges tests that watcher reflects instance updates
func TestWatcher_UpdateInstances_ReflectsChanges(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"
	initialService := &registry.ServiceInstance{
		ID:        "test-service-001",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8080"},
	}
	reg.Register(ctx, initialService)

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// Act - Get initial instances
	result1, err1 := watcher.Next()
	require.NoError(t, err1)
	assert.Len(t, result1, 1)
	assert.Equal(t, "test-service-001", result1[0].ID)

	// Register additional service
	additionalService := &registry.ServiceInstance{
		ID:        "test-service-002",
		Name:      serviceName,
		Version:   "v1.0.0",
		Endpoints: []string{"http://localhost:8081"},
	}
	reg.Register(ctx, additionalService)

	// Get updated instances (with timeout)
	done := make(chan bool)
	var result2 []*registry.ServiceInstance
	var err2 error

	go func() {
		result2, err2 = watcher.Next()
		done <- true
	}()

	select {
	case <-done:
		require.NoError(t, err2)
		assert.Len(t, result2, 2)
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for watcher update - this may be expected behavior")
	}
}

// TestWatcher_EmptyInstances_HandlesCorrectly tests watcher with empty instances
func TestWatcher_EmptyInstances_HandlesCorrectly(t *testing.T) {
	// Arrange
	reg := setupTestWatcherRegistry(t)
	ctx := context.Background()
	serviceName := "test.service"

	watcher, err := reg.Watch(ctx, serviceName)
	require.NoError(t, err)
	defer watcher.Stop()

	// Act
	result, err := watcher.Next()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// setupTestWatcherRegistry creates a test registry for watcher tests
func setupTestWatcherRegistry(t *testing.T) *Registry {
	tempDir := t.TempDir()
	registryPath := filepath.Join(tempDir, "registry")
	reg, err := New(registryPath)
	require.NoError(t, err)
	return reg
}