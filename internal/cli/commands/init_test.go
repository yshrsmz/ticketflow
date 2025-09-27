package commands

import (
	"context"
	flag "github.com/spf13/pflag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yshrsmz/ticketflow/internal/testutil"
)

func TestInitCommand(t *testing.T) {
	t.Run("command metadata", func(t *testing.T) {
		cmd := NewInitCommand()

		assert.Equal(t, "init", cmd.Name())
		assert.Nil(t, cmd.Aliases())
		assert.Equal(t, "Initialize a new ticketflow project", cmd.Description())
		assert.Equal(t, "init", cmd.Usage())
	})

	t.Run("no flags", func(t *testing.T) {
		cmd := NewInitCommand()

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		flags := cmd.SetupFlags(fs)

		assert.Nil(t, flags)
		assert.Equal(t, 0, fs.NFlag())
	})

	t.Run("validation always succeeds", func(t *testing.T) {
		cmd := NewInitCommand()

		err := cmd.Validate(nil, []string{})
		assert.NoError(t, err)

		err = cmd.Validate(nil, []string{"extra", "args"})
		assert.NoError(t, err)
	})

	t.Run("execute creates ticketflow structure", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir, err := os.MkdirTemp("", "ticketflow-init-test-*")
		require.NoError(t, err)
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Logf("failed to remove temp dir %s: %v", tmpDir, err)
			}
		}()

		// Change to temp directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		// Initialize git repo
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tmpDir
		require.NoError(t, gitInit.Run())
		testutil.GitConfigApply(t, testutil.SimpleGitExecutor{Dir: tmpDir})

		// Execute the init command
		cmd := NewInitCommand()
		err = cmd.Execute(context.Background(), nil, []string{})
		require.NoError(t, err)

		// Verify the structure was created
		assert.FileExists(t, filepath.Join(tmpDir, ".ticketflow.yaml"))
		assert.DirExists(t, filepath.Join(tmpDir, "tickets"))
		assert.DirExists(t, filepath.Join(tmpDir, "tickets", "todo"))
		assert.DirExists(t, filepath.Join(tmpDir, "tickets", "doing"))
		assert.DirExists(t, filepath.Join(tmpDir, "tickets", "done"))

		// Verify .gitignore was updated
		gitignoreContent, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
		require.NoError(t, err)
		assert.Contains(t, string(gitignoreContent), "current-ticket.md")
		assert.Contains(t, string(gitignoreContent), ".worktrees/")
	})

	t.Run("execute handles already initialized", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir, err := os.MkdirTemp("", "ticketflow-init-test-already-*")
		require.NoError(t, err)
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Logf("failed to remove temp dir %s: %v", tmpDir, err)
			}
		}()

		// Change to temp directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		// Initialize git repo
		gitInit := exec.Command("git", "init")
		gitInit.Dir = tmpDir
		require.NoError(t, gitInit.Run())
		testutil.GitConfigApply(t, testutil.SimpleGitExecutor{Dir: tmpDir})

		// Execute the init command first time
		cmd := NewInitCommand()
		err = cmd.Execute(context.Background(), nil, []string{})
		require.NoError(t, err)

		// Execute the init command second time
		err = cmd.Execute(context.Background(), nil, []string{})
		// Should succeed without error even when already initialized
		assert.NoError(t, err)
	})

	t.Run("execute fails when not in git repo", func(t *testing.T) {
		// Create a temporary directory for testing (without git init)
		tmpDir, err := os.MkdirTemp("", "ticketflow-init-test-no-git-*")
		require.NoError(t, err)
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Logf("failed to remove temp dir %s: %v", tmpDir, err)
			}
		}()

		// Change to temp directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		// Execute the init command without git repo
		cmd := NewInitCommand()
		err = cmd.Execute(context.Background(), nil, []string{})

		// Should fail as not in a git repository
		assert.Error(t, err)
	})
}
