package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default config",
			config: Config{
				Level:  "info",
				Format: "text",
				Output: "stderr",
			},
			wantErr: false,
		},
		{
			name: "json format",
			config: Config{
				Level:  "debug",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "invalid output file",
			config: Config{
				Level:  "info",
				Format: "text",
				Output: "/invalid/path/to/file.log",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLoggerWithMethods(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	// Test WithTicket
	ticketLogger := logger.WithTicket("test-ticket-123")
	ticketLogger.Info("ticket operation")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "test-ticket-123", logEntry["ticket_id"])
	assert.Equal(t, "ticket operation", logEntry["msg"])

	// Test WithOperation
	buf.Reset()
	opLogger := logger.WithOperation("create")
	opLogger.Info("creating resource")

	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, "create", logEntry["operation"])
	assert.Equal(t, "creating resource", logEntry["msg"])

	// Test WithError
	buf.Reset()
	testErr := assert.AnError
	errLogger := logger.WithError(testErr)
	errLogger.Error("operation failed")

	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, testErr.Error(), logEntry["error"])
	assert.Equal(t, "operation failed", logEntry["msg"])
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	logger.Info("test message", slog.String("key", "value"))

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	// Debug should not appear (below info level)
	logger.Debug("debug message")
	assert.Empty(t, buf.String())

	// Info should appear
	logger.Info("info message")
	assert.Contains(t, buf.String(), "info message")

	// Warn should appear
	buf.Reset()
	logger.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")

	// Error should appear
	buf.Reset()
	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestGlobalLogger(t *testing.T) {
	// Test default global logger
	assert.NotNil(t, Global())

	// Test setting global logger
	var buf bytes.Buffer
	customLogger := &Logger{
		slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	SetGlobal(customLogger)
	assert.Equal(t, customLogger, Global())

	// Test global convenience functions
	Info("global info", slog.String("test", "value"))

	output := buf.String()
	assert.Contains(t, output, "global info")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "value")
}
