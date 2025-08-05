package testutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSONOutput asserts that output is valid JSON and contains expected fields
func AssertJSONOutput(t *testing.T, output string, expectedFields ...string) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "Output should be valid JSON")

	for _, field := range expectedFields {
		assert.Contains(t, result, field, "JSON output should contain field: %s", field)
	}

	return result
}

// AssertJSONArray asserts that output is a valid JSON array
func AssertJSONArray(t *testing.T, output string) []interface{} {
	t.Helper()

	var result []interface{}
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "Output should be valid JSON array")

	return result
}

// AssertErrorContains asserts that an error contains a specific substring
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	require.Error(t, err, "Expected an error")
	assert.Contains(t, err.Error(), substr, "Error message should contain: %s", substr)
}

// AssertNoError is a helper that fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// AssertError is a helper that fails the test if err is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
}

// AssertOutputContains asserts that output contains expected strings
func AssertOutputContains(t *testing.T, output string, expected ...string) {
	t.Helper()
	for _, exp := range expected {
		assert.Contains(t, output, exp, "Output should contain: %s", exp)
	}
}

// AssertOutputNotContains asserts that output does not contain strings
func AssertOutputNotContains(t *testing.T, output string, notExpected ...string) {
	t.Helper()
	for _, exp := range notExpected {
		assert.NotContains(t, output, exp, "Output should not contain: %s", exp)
	}
}

// AssertOutputLines asserts specific lines in output
func AssertOutputLines(t *testing.T, output string, expectedLines ...string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, expected := range expectedLines {
		if i >= len(lines) {
			t.Errorf("Expected line %d to be %q, but output only has %d lines", i, expected, len(lines))
			continue
		}
		assert.Equal(t, expected, lines[i], "Line %d mismatch", i)
	}
}

// AssertOutputEmpty asserts that output is empty or only whitespace
func AssertOutputEmpty(t *testing.T, output string) {
	t.Helper()
	trimmed := strings.TrimSpace(output)
	assert.Empty(t, trimmed, "Output should be empty")
}

// AssertOutputJSON asserts that output is valid JSON and matches expected structure
func AssertOutputJSON(t *testing.T, output string, expected interface{}) {
	t.Helper()

	var actual interface{}
	err := json.Unmarshal([]byte(output), &actual)
	require.NoError(t, err, "Output should be valid JSON")

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err, "Failed to marshal expected value")

	var expectedParsed interface{}
	err = json.Unmarshal(expectedJSON, &expectedParsed)
	require.NoError(t, err, "Failed to unmarshal expected JSON")

	assert.Equal(t, expectedParsed, actual, "JSON output mismatch")
}

// AssertTicketFields asserts common ticket fields in a map
func AssertTicketFields(t *testing.T, ticket map[string]interface{}, expectedID string) {
	t.Helper()

	assert.Equal(t, expectedID, ticket["id"], "Ticket ID mismatch")
	assert.Contains(t, ticket, "priority", "Ticket should have priority")
	assert.Contains(t, ticket, "description", "Ticket should have description")
	assert.Contains(t, ticket, "created_at", "Ticket should have created_at")
	assert.Contains(t, ticket, "status", "Ticket should have status")
}

// AssertSliceContains asserts that a slice contains an element
func AssertSliceContains(t *testing.T, slice []string, element string) {
	t.Helper()
	assert.Contains(t, slice, element, "Slice should contain: %s", element)
}

// AssertSliceNotContains asserts that a slice does not contain an element
func AssertSliceNotContains(t *testing.T, slice []string, element string) {
	t.Helper()
	assert.NotContains(t, slice, element, "Slice should not contain: %s", element)
}

// AssertMapEqual asserts that two maps are equal
func AssertMapEqual(t *testing.T, expected, actual map[string]interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual, "Maps should be equal")
}

// AssertStringEqual is a simple string equality assertion
func AssertStringEqual(t *testing.T, expected, actual string) {
	t.Helper()
	assert.Equal(t, expected, actual, "Strings should be equal")
}

// AssertIntEqual is a simple int equality assertion
func AssertIntEqual(t *testing.T, expected, actual int) {
	t.Helper()
	assert.Equal(t, expected, actual, "Integers should be equal")
}

// AssertPanic asserts that a function panics
func AssertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected function to panic")
		}
	}()
	fn()
}

// AssertNoPanic asserts that a function does not panic
func AssertNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function panicked: %v", r)
		}
	}()
	fn()
}
