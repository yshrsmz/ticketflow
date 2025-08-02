// Package log provides structured logging for the ticketflow application.
package log

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger to provide application-specific logging methods.
type Logger struct {
	*slog.Logger
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
	level := parseLevel(cfg.Level)

	var output io.Writer
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

// WithContext returns a logger with context values.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger,
	}
}

// WithTicket returns a logger with ticket information.
func (l *Logger) WithTicket(ticketID string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("ticket_id", ticketID)),
	}
}

// WithOperation returns a logger with operation information.
func (l *Logger) WithOperation(op string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("operation", op)),
	}
}

// WithError returns a logger with error information.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(slog.String("error", err.Error())),
	}
}

// parseLevel converts string level to slog.Level.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
