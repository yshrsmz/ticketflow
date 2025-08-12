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
		assert.Equal(t, "version", cmd.Usage())
	})

	t.Run("no flags", func(t *testing.T) {
		cmd := NewVersionCommand("1.2.3", "abc123", "2024-01-01")

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs)

		assert.Nil(t, flags)
		assert.Equal(t, 0, fs.NFlag())
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

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
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

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "ticketflow version dev")
		// Should not show commit and build time for dev version with unknown values
		assert.False(t, strings.Contains(output, "Git commit"))
		assert.False(t, strings.Contains(output, "Built at"))
	})
}