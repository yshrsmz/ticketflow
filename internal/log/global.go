package log

import (
	"sync"
)

var (
	globalLogger *Logger
	globalMu     sync.RWMutex
)

func init() {
	// Initialize with no-op logger (silent by default)
	globalLogger = NewNoOp()
}

// SetGlobal sets the global logger instance.
func SetGlobal(logger *Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}

// Global returns the global logger instance.
func Global() *Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

// Debug logs at debug level using the global logger.
func Debug(msg string, args ...any) {
	Global().Debug(msg, args...)
}

// Info logs at info level using the global logger.
func Info(msg string, args ...any) {
	Global().Info(msg, args...)
}

// Warn logs at warn level using the global logger.
func Warn(msg string, args ...any) {
	Global().Warn(msg, args...)
}

// Error logs at error level using the global logger.
func Error(msg string, args ...any) {
	Global().Error(msg, args...)
}
