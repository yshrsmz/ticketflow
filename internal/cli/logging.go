package cli

import (
	"flag"
	"fmt"

	"github.com/yshrsmz/ticketflow/internal/log"
)

// LoggingOptions holds command-line logging configuration
type LoggingOptions struct {
	Level  string
	Format string
	Output string
}

// AddLoggingFlags adds logging-related flags to a flag set
func AddLoggingFlags(fs *flag.FlagSet) *LoggingOptions {
	opts := &LoggingOptions{}

	fs.StringVar(&opts.Level, "log-level", "", "Log level (debug, info, warn, error)")
	fs.StringVar(&opts.Format, "log-format", "", "Log format (text, json)")
	fs.StringVar(&opts.Output, "log-output", "", "Log output (stderr, stdout, or file path)")

	return opts
}

// ConfigureLogging sets up logging based on command-line options
func ConfigureLogging(opts *LoggingOptions) error {
	// If no logging options provided, keep the default no-op logger
	if opts.Level == "" && opts.Format == "" && opts.Output == "" {
		return nil
	}

	// Build config from options
	cfg := log.Config{
		Level:  opts.Level,
		Format: opts.Format,
		Output: opts.Output,
	}

	// Apply defaults for any unspecified options
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	if cfg.Output == "" {
		cfg.Output = "stderr"
	}

	// Create and set the logger
	logger, err := log.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to configure logging: %w", err)
	}

	log.SetGlobal(logger)
	return nil
}
