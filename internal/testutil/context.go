package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CancelledContext returns a context that has already been cancelled
func CancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TimeoutContext returns a context with the specified timeout
func TimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// ShortTimeoutContext returns a context with a very short timeout for testing
func ShortTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 50*time.Millisecond)
}

// AssertContextError asserts that an error is a context cancellation error
func AssertContextError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err, "Expected context error")
	assert.Contains(t, err.Error(), "context canceled", "Error should be context cancellation")
}

// AssertTimeoutError asserts that an error is a context timeout error
func AssertTimeoutError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err, "Expected timeout error")
	assert.Contains(t, err.Error(), "context deadline exceeded", "Error should be context timeout")
}

// RunWithTimeout runs a function with a timeout and returns any error
func RunWithTimeout(t *testing.T, timeout time.Duration, fn func(context.Context) error) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return fn(ctx)
}

// WaitForContext waits for a context to be done or a timeout
func WaitForContext(t *testing.T, ctx context.Context, timeout time.Duration) bool {
	t.Helper()
	select {
	case <-ctx.Done():
		return true
	case <-time.After(timeout):
		return false
	}
}

// ContextWithValue creates a context with a test value
func ContextWithValue(key, value interface{}) context.Context {
	return context.WithValue(context.Background(), key, value)
}

// BlockingOperation simulates a blocking operation that respects context
func BlockingOperation(ctx context.Context, blockTime time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(blockTime):
		return nil
	}
}

// TestContextKey is a key type for test context values
type TestContextKey string

const (
	// TestIDKey is the context key for test ID
	TestIDKey TestContextKey = "test-id"
	// TestUserKey is the context key for test user
	TestUserKey TestContextKey = "test-user"
)

// WithTestID adds a test ID to the context
func WithTestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TestIDKey, id)
}

// GetTestID retrieves the test ID from context
func GetTestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(TestIDKey).(string)
	return id, ok
}

// WithTestUser adds a test user to the context
func WithTestUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, TestUserKey, user)
}

// GetTestUser retrieves the test user from context
func GetTestUser(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(TestUserKey).(string)
	return user, ok
}
