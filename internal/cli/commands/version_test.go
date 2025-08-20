package commands

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	t.Run("command metadata", func(t *testing.T) {
		cmd := NewVersionCommand("1.2.3", "abc123", "2024-01-01")

		assert.Equal(t, "version", cmd.Name())
		assert.Equal(t, []string{"-v", "--version"}, cmd.Aliases())
		assert.Equal(t, "Show version information", cmd.Description())
		assert.Equal(t, "version [--format text|json]", cmd.Usage())
	})

	t.Run("has format flag", func(t *testing.T) {
		cmd := NewVersionCommand("1.2.3", "abc123", "2024-01-01")

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs)

		assert.NotNil(t, flags)
		// NFlag returns number of flags that have been set, not defined
		// Since we just defined it, not set it, NFlag should be 0

		// Check that format flag exists
		formatFlag := fs.Lookup("format")
		assert.NotNil(t, formatFlag)
		assert.Equal(t, "text", formatFlag.DefValue)
	})

	t.Run("validation always succeeds", func(t *testing.T) {
		cmd := NewVersionCommand("1.2.3", "abc123", "2024-01-01")

		err := cmd.Validate(nil, []string{})
		assert.NoError(t, err)

		err = cmd.Validate(nil, []string{"extra", "args"})
		assert.NoError(t, err)
	})

	t.Run("execute with version info", func(t *testing.T) {
		cmd := NewVersionCommand("1.2.3", "abc123", "2024-01-01")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Execute(context.Background(), nil, []string{})
		require.NoError(t, err)

		require.NoError(t, w.Close())
		os.Stdout = old

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "ticketflow version 1.2.3")
		assert.Contains(t, output, "Git commit: abc123")
		assert.Contains(t, output, "Built at:   2024-01-01")
	})

	t.Run("execute with dev version", func(t *testing.T) {
		cmd := NewVersionCommand("dev", "unknown", "unknown")

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Execute(context.Background(), nil, []string{})
		require.NoError(t, err)

		require.NoError(t, w.Close())
		os.Stdout = old

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		output := buf.String()

		assert.Contains(t, output, "ticketflow version dev")
		// Should not show commit and build time for dev version with unknown values
		assert.False(t, strings.Contains(output, "Git commit"))
		assert.False(t, strings.Contains(output, "Built at"))
	})
}
