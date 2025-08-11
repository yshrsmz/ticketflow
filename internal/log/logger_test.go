package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
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
		{
			name: "invalid log level",
			config: Config{
				Level:  "invalid-level",
				Format: "text",
				Output: "stderr",
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
	t.Parallel()
	var buf bytes.Buffer

	logger := &Logger{
		Logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		closer: nil,
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
	// The error is now stored as an object, not just a string
	assert.NotNil(t, logEntry["error"])
	assert.Equal(t, "operation failed", logEntry["msg"])
}

func TestParseLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input     string
		expected  slog.Level
		wantError bool
	}{
		{"debug", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"", slog.LevelInfo, false}, // empty is allowed
		{"invalid", slog.LevelInfo, true},
		{"unknown-level", slog.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid log level")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestTextFormat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	logger := &Logger{
		Logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		closer: nil,
	}

	logger.Info("test message", slog.String("key", "value"))

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestLogLevels(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	logger := &Logger{
		Logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		closer: nil,
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
	t.Parallel()
	// Test default global logger
	assert.NotNil(t, Global())

	// Test setting global logger
	var buf bytes.Buffer
	customLogger := &Logger{
		Logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		closer: nil,
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

func TestClose(t *testing.T) {
	t.Parallel()
	// Test Close with no closer (stdout/stderr)
	logger := &Logger{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		closer: nil,
	}
	err := logger.Close()
	assert.NoError(t, err)

	// Test Close with a mock closer
	mockCloser := &mockCloser{closed: false}
	logger = &Logger{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		closer: mockCloser,
	}
	err = logger.Close()
	assert.NoError(t, err)
	assert.True(t, mockCloser.closed)
}

type mockCloser struct {
	closed bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}
