package cli

import (
	"fmt"
	"io"
)

// StatusWriter handles progress and status messages during command execution.
// In text mode, messages are displayed to the user.
// In JSON mode, messages are suppressed to maintain valid JSON output.
type StatusWriter interface {
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// textStatusWriter outputs status messages to the console in text mode
type textStatusWriter struct {
	w io.Writer
}

// NewTextStatusWriter creates a status writer that outputs to the given writer
func NewTextStatusWriter(w io.Writer) StatusWriter {
	return &textStatusWriter{w: w}
}

func (s *textStatusWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(s.w, format, args...)
}

func (s *textStatusWriter) Println(args ...interface{}) {
	fmt.Fprintln(s.w, args...)
}

// nullStatusWriter suppresses all status messages (used for JSON mode)
type nullStatusWriter struct{}

// NewNullStatusWriter creates a status writer that suppresses all output
func NewNullStatusWriter() StatusWriter {
	return &nullStatusWriter{}
}

func (s *nullStatusWriter) Printf(format string, args ...interface{}) {
	// No-op: suppress output in JSON mode
}

func (s *nullStatusWriter) Println(args ...interface{}) {
	// No-op: suppress output in JSON mode
}

// NewStatusWriter creates the appropriate status writer based on the output format
func NewStatusWriter(w io.Writer, format OutputFormat) StatusWriter {
	if format == FormatJSON {
		return NewNullStatusWriter()
	}
	return NewTextStatusWriter(w)
}