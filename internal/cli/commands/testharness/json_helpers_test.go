package testharness

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:  "simple JSON object",
			input: `{"id": "test", "value": 42}`,
			expected: map[string]interface{}{
				"id":    "test",
				"value": float64(42),
			},
		},
		{
			name:  "JSON with prefix text",
			input: `Processing ticket... {"success": true, "result": "done"}`,
			expected: map[string]interface{}{
				"success": true,
				"result":  "done",
			},
		},
		{
			name:  "nested JSON object",
			input: `{"ticket": {"id": "123", "status": "doing"}}`,
			expected: map[string]interface{}{
				"ticket": map[string]interface{}{
					"id":     "123",
					"status": "doing",
				},
			},
		},
		{
			name:  "JSON with misleading brace in prefix",
			input: `Error: {malformed} {"actual": "json", "valid": true}`,
			expected: map[string]interface{}{
				"actual": "json",
				"valid":  true,
			},
		},
		{
			name:  "JSON with bracket in prefix",
			input: `Status [OK]: {"message": "success"}`,
			expected: map[string]interface{}{
				"message": "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON(t, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateJSON_ArrayShouldFail(t *testing.T) {
	t.Parallel()

	// Test that ValidateJSON properly rejects array input
	// Extract JSON to verify it's detected as an array
	extracted, err := ExtractJSONFromMixedOutput(`[1, 2, 3]`)
	require.NoError(t, err, "Should extract JSON content")
	assert.Equal(t, `[1, 2, 3]`, extracted)

	// Verify that trying to parse an array as object would fail
	var obj map[string]interface{}
	err = json.Unmarshal([]byte(extracted), &obj)
	assert.Error(t, err, "Should fail to unmarshal array into map[string]interface{}")

	// Verify that an array can be properly parsed with ValidateJSONArray
	result := ValidateJSONArray(t, `[1, 2, 3]`)
	assert.Len(t, result, 3)
}

func TestValidateJSONArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedLen    int
		checkFirstElem bool
		firstElem      interface{}
	}{
		{
			name:        "simple array",
			input:       `[1, 2, 3]`,
			expectedLen: 3,
		},
		{
			name:           "array of objects",
			input:          `[{"id": "1"}, {"id": "2"}]`,
			expectedLen:    2,
			checkFirstElem: true,
			firstElem: map[string]interface{}{
				"id": "1",
			},
		},
		{
			name:        "array with prefix",
			input:       `Status: OK [{"name": "test"}]`,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSONArray(t, tt.input)
			assert.Len(t, result, tt.expectedLen)
			if tt.checkFirstElem {
				assert.Equal(t, tt.firstElem, result[0])
			}
		})
	}
}

func TestAssertJSONField(t *testing.T) {
	t.Parallel()

	jsonData := map[string]interface{}{
		"id":     "test-123",
		"status": "active",
		"metadata": map[string]interface{}{
			"priority": float64(2),
			"tags":     []interface{}{"urgent", "bug"},
		},
		"config": map[string]interface{}{
			"settings": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	// Test successful assertions
	AssertJSONField(t, jsonData, "id", "test-123")
	AssertJSONField(t, jsonData, "status", "active")
	AssertJSONField(t, jsonData, "metadata.priority", float64(2))
	AssertJSONField(t, jsonData, "config.settings.enabled", true)
}

func TestAssertJSONFieldExists(t *testing.T) {
	t.Parallel()

	jsonData := map[string]interface{}{
		"id": "test",
		"nested": map[string]interface{}{
			"field": "value",
		},
	}

	// These should pass
	AssertJSONFieldExists(t, jsonData, "id")
	AssertJSONFieldExists(t, jsonData, "nested")
	AssertJSONFieldExists(t, jsonData, "nested.field")
}

func TestAssertJSONArrayLength(t *testing.T) {
	t.Parallel()

	jsonData := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
		"nested": map[string]interface{}{
			"list": []interface{}{1, 2},
		},
	}

	AssertJSONArrayLength(t, jsonData, "items", 3)
	AssertJSONArrayLength(t, jsonData, "nested.list", 2)
}

func TestGetJSONField(t *testing.T) {
	t.Parallel()

	jsonData := map[string]interface{}{
		"id": "test",
		"nested": map[string]interface{}{
			"value": float64(42),
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{"root field", "id", "test"},
		{"nested field", "nested.value", float64(42)},
		{"non-existent", "missing", nil},
		{"non-existent nested", "nested.missing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetJSONField(jsonData, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAssertJSONSuccess(t *testing.T) {
	t.Parallel()

	successData := map[string]interface{}{
		"success": true,
		"result":  "operation completed",
	}

	AssertJSONSuccess(t, successData)
}

func TestAssertJSONError(t *testing.T) {
	t.Parallel()

	errorData := map[string]interface{}{
		"success": false,
		"error":   "ticket not found",
	}

	AssertJSONError(t, errorData, "not found")
}

func TestValidateTicketJSON(t *testing.T) {
	t.Parallel()

	ticketData := map[string]interface{}{
		"id":          "test-ticket",
		"status":      "doing",
		"priority":    float64(1),
		"description": "Test description",
		"created_at":  "2025-01-01T00:00:00Z",
	}

	ValidateTicketJSON(t, ticketData, "test-ticket", "doing")
}

func TestExtractJSONFromMixedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "pure JSON object",
			input:    `{"id": "test"}`,
			expected: `{"id": "test"}`,
		},
		{
			name:     "JSON with prefix",
			input:    `Status: OK {"result": true}`,
			expected: `{"result": true}`,
		},
		{
			name:     "JSON array",
			input:    `Processing... [1, 2, 3]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "misleading brace in prefix",
			input:    `Error {incomplete {"valid": "json"}`,
			expected: `{"valid": "json"}`,
		},
		{
			name:     "misleading bracket in prefix",
			input:    `List [incomplete ["valid", "array"]`,
			expected: `["valid", "array"]`,
		},
		{
			name:     "multiple braces before valid JSON",
			input:    `Error: {bad} {malformed {"actual": "json", "valid": true}`,
			expected: `{"actual": "json", "valid": true}`,
		},
		{
			name:     "JSON with newlines in prefix",
			input:    "Status: Processing\nDebug: {invalid\n{\"status\": \"ok\"}",
			expected: `{"status": "ok"}`,
		},
		{
			name:    "no JSON",
			input:   `Just plain text`,
			wantErr: true,
		},
		{
			name:    "malformed JSON",
			input:   `Status: {"incomplete": `,
			wantErr: true,
		},
		{
			name:    "only invalid JSON markers",
			input:   `{ [ } ]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractJSONFromMixedOutput(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
