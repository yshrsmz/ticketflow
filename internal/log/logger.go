// Package log provides structured logging for the ticketflow application.
package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger to provide application-specific logging methods.
type Logger struct {
	*slog.Logger
	closer io.Closer // For closing file outputs
}

// Config holds logging configuration.
type Config struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // text, json
	Output string `yaml:"output"` // stderr, stdout, or file path
}

// DefaultConfig returns the default logging configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Format: "text",
		Output: "stderr",
	}
}

// New creates a new logger with the given configuration.
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	var output io.Writer
	var closer io.Closer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr", "":
		output = os.Stderr
	default:
		// File path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
		closer = file
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text", "":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		closer: closer,
	}, nil
}

// Default creates a no-op logger that discards all output.
func Default() *Logger {
	return NewNoOp()
}

// NewNoOp creates a logger that discards all output (silent operation).
func NewNoOp() *Logger {
	// Use io.Discard to throw away all log output
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.Level(1000), // Set an impossibly high level so nothing gets logged
	})
	return &Logger{
		Logger: slog.New(handler),
	}
}

// Close closes the logger's output if it's a file.
func (l *Logger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// WithTicket returns a logger with ticket information.
func (l *Logger) WithTicket(ticketID string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("ticket_id", ticketID)),
		closer: l.closer,
	}
}

// WithOperation returns a logger with operation information.
func (l *Logger) WithOperation(op string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("operation", op)),
		closer: l.closer,
	}
}

// WithError returns a logger with error information.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(slog.Any("error", err)),
		closer: l.closer,
	}
}

// parseLevel converts string level to slog.Level.
func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "": // Allow empty for default
		return slog.LevelInfo, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
