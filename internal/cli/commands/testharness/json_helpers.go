// Package testharness provides integration testing utilities for CLI commands.
// The JSON validation helpers are designed to be simple, explicit, and debuggable
// without external dependencies, focusing on practical test scenarios.
package testharness

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ValidateJSON parses JSON object output and returns the unmarshaled structure.
// It strips any non-JSON content (status messages, etc.) before parsing.
// Note: This function only handles JSON objects. For JSON arrays, use ValidateJSONArray.
func ValidateJSON(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	// Extract JSON content, looking specifically for an object
	jsonStr, isObject := extractJSONContent(output)
	require.True(t, isObject, "Expected JSON object but got array or no JSON. Use ValidateJSONArray for arrays.")
	require.NotEmpty(t, jsonStr, "No JSON content found in output")

	// Parse JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err, "Failed to parse JSON output: %s", jsonStr)

	return result
}

// ValidateJSONArray parses JSON array output and returns the unmarshaled structure.
// It strips any non-JSON content (status messages, etc.) before parsing.
// Note: This function only handles JSON arrays. For JSON objects, use ValidateJSON.
func ValidateJSONArray(t *testing.T, output string) []interface{} {
	t.Helper()

	// Extract JSON content, looking specifically for an array
	jsonStr, isObject := extractJSONContent(output)
	require.False(t, isObject, "Expected JSON array but got object. Use ValidateJSON for objects.")
	require.NotEmpty(t, jsonStr, "No JSON array content found in output")

	// Parse JSON
	var result []interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err, "Failed to parse JSON array output: %s", jsonStr)

	return result
}

// AssertJSONField validates a specific field in JSON output.
// The path can be dot-separated for nested fields (e.g., "current_ticket.id").
// The expected value is compared using testify's assert.Equal.
func AssertJSONField(t *testing.T, jsonData map[string]interface{}, path string, expected interface{}) {
	t.Helper()

	parts := strings.Split(path, ".")
	current := jsonData

	// Navigate nested structure
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - check the value
			actual, exists := current[part]
			require.True(t, exists, "Field %q not found in JSON (path: %q)", part, path)
			assert.Equal(t, expected, actual, "Field %q has unexpected value", path)
			return
		}

		// Intermediate part - navigate deeper
		value, exists := current[part]
		require.True(t, exists, "Field %q not found in JSON (path: %q)", part, path)

		// Check if it's a map we can navigate into
		nested, ok := value.(map[string]interface{})
		require.True(t, ok, "Field %q is not a nested object (path: %q)", part, path)
		current = nested
	}
}

// AssertJSONFieldExists checks that a field exists in the JSON output without validating its value.
// The path can be dot-separated for nested fields (e.g., "current_ticket.id").
func AssertJSONFieldExists(t *testing.T, jsonData map[string]interface{}, path string) {
	t.Helper()

	parts := strings.Split(path, ".")
	current := jsonData

	// Navigate nested structure
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - check existence
			_, exists := current[part]
			assert.True(t, exists, "Field %q not found in JSON (path: %q)", part, path)
			return
		}

		// Intermediate part - navigate deeper
		value, exists := current[part]
		require.True(t, exists, "Field %q not found in JSON (path: %q)", part, path)

		// Check if it's a map we can navigate into
		nested, ok := value.(map[string]interface{})
		require.True(t, ok, "Field %q is not a nested object (path: %q)", part, path)
		current = nested
	}
}

// AssertJSONArrayLength validates the length of an array field in JSON output.
func AssertJSONArrayLength(t *testing.T, jsonData map[string]interface{}, path string, expectedLength int) {
	t.Helper()

	parts := strings.Split(path, ".")
	current := jsonData

	// Navigate to the array field
	for i, part := range parts {
		value, exists := current[part]
		require.True(t, exists, "Field %q not found in JSON (path: %q)", part, path)

		if i == len(parts)-1 {
			// Last part - should be an array
			arr, ok := value.([]interface{})
			require.True(t, ok, "Field %q is not an array (path: %q)", part, path)
			assert.Len(t, arr, expectedLength, "Array field %q has unexpected length", path)
			return
		}

		// Intermediate part - navigate deeper
		nested, ok := value.(map[string]interface{})
		require.True(t, ok, "Field %q is not a nested object (path: %q)", part, path)
		current = nested
	}
}

// GetJSONField retrieves a field value from JSON data using a dot-separated path.
// Returns nil if the field doesn't exist.
func GetJSONField(jsonData map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := jsonData

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil
		}

		if i == len(parts)-1 {
			// Last part - return the value
			return value
		}

		// Intermediate part - navigate deeper
		nested, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		current = nested
	}

	return nil
}

// AssertJSONSuccess validates that a JSON response has success=true and optional result data.
func AssertJSONSuccess(t *testing.T, jsonData map[string]interface{}) {
	t.Helper()
	AssertJSONField(t, jsonData, "success", true)
}

// AssertJSONError validates that a JSON response has success=false and contains an error message.
func AssertJSONError(t *testing.T, jsonData map[string]interface{}, errorContains string) {
	t.Helper()
	AssertJSONField(t, jsonData, "success", false)

	errorMsg, exists := jsonData["error"]
	require.True(t, exists, "Error field not found in JSON error response")

	errorStr, ok := errorMsg.(string)
	require.True(t, ok, "Error field is not a string")

	if errorContains != "" {
		assert.Contains(t, errorStr, errorContains, "Error message doesn't contain expected text")
	}
}

// ValidateTicketJSON validates common ticket fields in JSON output.
func ValidateTicketJSON(t *testing.T, ticketData map[string]interface{}, expectedID, expectedStatus string) {
	t.Helper()

	// Validate required fields exist
	AssertJSONFieldExists(t, ticketData, "id")
	AssertJSONFieldExists(t, ticketData, "status")
	AssertJSONFieldExists(t, ticketData, "priority")
	AssertJSONFieldExists(t, ticketData, "description")
	AssertJSONFieldExists(t, ticketData, "created_at")

	// Validate specific values if provided
	if expectedID != "" {
		AssertJSONField(t, ticketData, "id", expectedID)
	}
	if expectedStatus != "" {
		AssertJSONField(t, ticketData, "status", expectedStatus)
	}
}

// ExtractJSONFromMixedOutput extracts JSON content from output that may contain
// non-JSON text (like status messages). Returns the JSON string portion.
func ExtractJSONFromMixedOutput(output string) (string, error) {
	jsonStr, _ := extractJSONContent(output)
	if jsonStr == "" {
		return "", fmt.Errorf("no JSON content found in output")
	}
	return jsonStr, nil
}

// extractJSONContent attempts to extract JSON content from mixed output.
// It uses a simple heuristic approach suitable for test scenarios where we control the output format.
// Returns the extracted JSON string and whether it's an object (true) or array (false).
func extractJSONContent(output string) (string, bool) {
	// Try each position in the string to find valid JSON start
	for i := 0; i < len(output); i++ {
		switch output[i] {
		case '{':
			// Potential object start - validate it's actual JSON
			jsonStr := output[i:]
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &obj); err == nil {
				return jsonStr, true
			}
		case '[':
			// Potential array start - validate it's actual JSON
			jsonStr := output[i:]
			var arr []interface{}
			if err := json.Unmarshal([]byte(jsonStr), &arr); err == nil {
				return jsonStr, false
			}
		}
	}
	return "", false
}
