// Package cli_test provides comprehensive testing for the CLI package.
// Tests are organized into:
// - Unit tests for individual functions
// - Integration tests for command workflows
// - Benchmarks for performance-critical paths
//
// Test files are organized by functionality:
// - errors_extended_test.go: Error handling and conversion tests
// - output_extended_test.go: Output formatting tests with benchmarks
// - cleanup_extended_test.go: Cleanup functionality tests
// - error_converter_test.go: Error conversion tests
// - test_helpers.go: Common test utilities and fixtures
//
// All tests follow Go best practices:
// - Table-driven tests for comprehensive coverage
// - Parallel execution where safe
// - Proper resource cleanup with defer
// - Mock usage for external dependencies
// - Fixed time values for deterministic tests
package cli
