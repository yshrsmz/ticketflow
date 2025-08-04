package testutil

import (
	"bytes"
	"io"
)

// OutputCapture helps capture output during tests
type OutputCapture struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

// NewOutputCapture creates a new output capture instance
func NewOutputCapture() *OutputCapture {
	return &OutputCapture{}
}

// Write implements io.Writer (for stdout)
func (c *OutputCapture) Write(p []byte) (n int, err error) {
	return c.stdout.Write(p)
}

// WriteString implements io.StringWriter (for stdout)
func (c *OutputCapture) WriteString(s string) (n int, err error) {
	return c.stdout.WriteString(s)
}

// Stdout returns the captured stdout
func (c *OutputCapture) Stdout() string {
	return c.stdout.String()
}

// Stderr returns the captured stderr
func (c *OutputCapture) Stderr() string {
	return c.stderr.String()
}

// StdoutWriter returns an io.Writer for stdout
func (c *OutputCapture) StdoutWriter() io.Writer {
	return &c.stdout
}

// StderrWriter returns an io.Writer for stderr
func (c *OutputCapture) StderrWriter() io.Writer {
	return &c.stderr
}

// Reset clears the captured output
func (c *OutputCapture) Reset() {
	c.stdout.Reset()
	c.stderr.Reset()
}
