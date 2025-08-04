package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWithContext(t *testing.T) {
	t.Parallel()
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ticketflow.yaml")

	// Save default config
	cfg := Default()
	err := cfg.Save(configPath)
	require.NoError(t, err)

	t.Run("successful load with context", func(t *testing.T) {
		ctx := context.Background()
		loaded, err := LoadWithContext(ctx, tmpDir)
		require.NoError(t, err)
		assert.Equal(t, cfg.Git.DefaultBranch, loaded.Git.DefaultBranch)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := LoadWithContext(ctx, tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})
}

func TestSaveWithContext(t *testing.T) {
	t.Parallel()
	// Create temp directory
	tmpDir := t.TempDir()

	t.Run("successful save with context", func(t *testing.T) {
		ctx := context.Background()
		cfg := Default()
		cfg.Git.DefaultBranch = "test-branch"

		configPath := filepath.Join(tmpDir, "config1.yaml")
		err := cfg.SaveWithContext(ctx, configPath)
		require.NoError(t, err)

		// Verify file was saved correctly by loading it back
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "test-branch")
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cfg := Default()
		configPath := filepath.Join(tmpDir, "config2.yaml")
		err := cfg.SaveWithContext(ctx, configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")

		// Verify file was not created
		_, err = os.Stat(configPath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestConfigFileContextHelpers(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("readConfigFileWithContext", func(t *testing.T) {
		// Create a test config file
		configPath := filepath.Join(tmpDir, "test-config.yaml")
		testData := []byte("test: config\ndata: value")
		err := os.WriteFile(configPath, testData, 0644)
		require.NoError(t, err)

		// Test successful read
		ctx := context.Background()
		data, err := readConfigFileWithContext(ctx, configPath)
		require.NoError(t, err)
		assert.Equal(t, testData, data)

		// Test with cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = readConfigFileWithContext(ctx, configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")
	})

	t.Run("writeConfigFileWithContext", func(t *testing.T) {
		testData := []byte("test: config\ndata: value")

		// Test successful write
		ctx := context.Background()
		configPath := filepath.Join(tmpDir, "write-test.yaml")
		err := writeConfigFileWithContext(ctx, configPath, testData, 0644)
		require.NoError(t, err)

		// Verify file was written
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, testData, data)

		// Test with cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		configPath2 := filepath.Join(tmpDir, "write-test2.yaml")
		err = writeConfigFileWithContext(ctx, configPath2, testData, 0644)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation cancelled")

		// Verify file was not created
		_, err = os.Stat(configPath2)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("atomic write behavior", func(t *testing.T) {
		// Test that write is atomic (uses temp file and rename)
		ctx := context.Background()
		configPath := filepath.Join(tmpDir, "atomic-test.yaml")

		// Write initial content
		initialData := []byte("initial: content")
		err := writeConfigFileWithContext(ctx, configPath, initialData, 0644)
		require.NoError(t, err)

		// Simulate partial write by creating a goroutine that tries to read
		// while we're writing new content
		newData := []byte("new: content\nmore: data")
		done := make(chan struct{})
		go func() {
			// Try to read file multiple times during write
			for i := 0; i < 10; i++ {
				data, _ := os.ReadFile(configPath)
				// Should always see either initial or new content, never partial
				if len(data) > 0 {
					assert.True(t, string(data) == string(initialData) || string(data) == string(newData))
				}
				time.Sleep(time.Millisecond)
			}
			close(done)
		}()

		// Write new content
		err = writeConfigFileWithContext(ctx, configPath, newData, 0644)
		require.NoError(t, err)

		<-done
	})

	t.Run("config file size limit", func(t *testing.T) {
		// Test that large config files are rejected
		ctx := context.Background()
		configPath := filepath.Join(tmpDir, "large-config.yaml")

		// Create a file larger than MaxConfigSize limit
		largeData := make([]byte, MaxConfigSize+1)
		for i := range largeData {
			largeData[i] = 'a'
		}
		err := os.WriteFile(configPath, largeData, 0644)
		require.NoError(t, err)

		// Try to read it
		_, err = readConfigFileWithContext(ctx, configPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config file too large")
	})
}
