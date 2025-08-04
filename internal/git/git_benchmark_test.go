package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkGitExec benchmarks basic git command execution
func BenchmarkGitExec(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	commands := []struct {
		name string
		args []string
	}{
		{"status", []string{"status", "--porcelain"}},
		{"branch-list", []string{"branch", "--list"}},
		{"log-oneline", []string{"log", "--oneline", "-n", "1"}},
	}

	for _, cmd := range commands {
		b.Run(cmd.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := git.Exec(ctx, cmd.args...)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCreateBranch benchmarks branch creation
func BenchmarkCreateBranch(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	// Cleanup created branches after benchmark
	b.Cleanup(func() {
		// Switch to main/master first
		_, _ = git.Exec(ctx, "checkout", "main")

		// Get all benchmark branches
		output, _ := git.Exec(ctx, "branch", "--list", "benchmark-branch-*")
		if output != "" {
			// Delete branches
			_, _ = git.Exec(ctx, "branch", "-D", "benchmark-branch-*")
		}
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		branchName := fmt.Sprintf("benchmark-branch-%d", i)
		err := git.CreateBranch(ctx, branchName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBranchExists benchmarks checking if branches exist
func BenchmarkBranchExists(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	// Create some branches
	for i := 0; i < 10; i++ {
		branchName := fmt.Sprintf("test-branch-%d", i)
		if err := git.CreateBranch(ctx, branchName); err != nil {
			b.Fatal(err)
		}
	}

	scenarios := []struct {
		name       string
		branchName string
		exists     bool
	}{
		{"existing-branch", "test-branch-5", true},
		{"non-existing-branch", "non-existent", false},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				exists, err := git.BranchExists(ctx, scenario.branchName)
				if err != nil {
					b.Fatal(err)
				}
				if exists != scenario.exists {
					b.Fatalf("expected %v, got %v", scenario.exists, exists)
				}
			}
		})
	}
}

// BenchmarkCurrentBranch benchmarks getting the current branch
func BenchmarkCurrentBranch(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := git.CurrentBranch(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListWorktrees benchmarks listing worktrees
func BenchmarkListWorktrees(b *testing.B) {
	scenarios := []struct {
		name          string
		worktreeCount int
	}{
		{"no-worktrees", 0},
		{"5-worktrees", 5},
		{"10-worktrees", 10},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			// Setup once before the benchmark loop
			tmpDir := b.TempDir()
			setupBenchmarkRepo(b, tmpDir)
			git := New(tmpDir)
			ctx := context.Background()

			// Only create worktrees if needed
			if scenario.worktreeCount > 0 {
				// Ensure we're on the main branch
				currentBranch, err := git.CurrentBranch(ctx)
				if err != nil {
					b.Fatal(err)
				}

				// Create worktrees outside the main repo directory to avoid conflicts
				parentDir := filepath.Dir(tmpDir)
				worktreesDir := filepath.Join(parentDir, "bench-worktrees")
				if err := os.MkdirAll(worktreesDir, 0755); err != nil {
					b.Fatal(err)
				}

				// Create worktrees with unique branch names
				for i := 0; i < scenario.worktreeCount; i++ {
					branchName := fmt.Sprintf("bench-wt-%d", i)
					worktreePath := filepath.Join(worktreesDir, branchName)

					// Create branch from current branch
					if _, err := git.Exec(ctx, "branch", branchName, currentBranch); err != nil {
						b.Fatal(err)
					}

					if err := git.AddWorktree(ctx, worktreePath, branchName); err != nil {
						b.Fatal(err)
					}
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := git.ListWorktrees(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkAddWorktree benchmarks worktree creation
func BenchmarkAddWorktree(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	// Pre-create branches for worktrees
	for i := 0; i < b.N; i++ {
		branchName := fmt.Sprintf("worktree-branch-%d", i)
		if err := git.CreateBranch(ctx, branchName); err != nil {
			b.Fatal(err)
		}
	}

	// Cleanup worktrees and branches after benchmark
	b.Cleanup(func() {
		// Remove all worktrees
		worktrees, _ := git.ListWorktrees(ctx)
		for _, wt := range worktrees {
			if wt.Path != tmpDir { // Don't remove main worktree
				_ = git.RemoveWorktree(ctx, wt.Path)
			}
		}

		// Switch to main and delete branches
		_, _ = git.Exec(ctx, "checkout", "main")
		_, _ = git.Exec(ctx, "branch", "-D", "worktree-branch-*")
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		branchName := fmt.Sprintf("worktree-branch-%d", i)
		worktreePath := filepath.Join(tmpDir, ".worktrees", branchName)
		err := git.AddWorktree(ctx, worktreePath, branchName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRemoveWorktree benchmarks worktree removal
func BenchmarkRemoveWorktree(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	// Pre-create worktrees to remove
	worktreePaths := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		branchName := fmt.Sprintf("worktree-branch-%d", i)
		worktreePath := filepath.Join(tmpDir, ".worktrees", branchName)
		if err := git.CreateBranch(ctx, branchName); err != nil {
			b.Fatal(err)
		}
		if err := git.AddWorktree(ctx, worktreePath, branchName); err != nil {
			b.Fatal(err)
		}
		worktreePaths[i] = worktreePath
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := git.RemoveWorktree(ctx, worktreePaths[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCommit benchmarks git commit operations
func BenchmarkCommit(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create a file to commit
		fileName := fmt.Sprintf("file-%d.txt", i)
		filePath := filepath.Join(tmpDir, fileName)
		content := fmt.Sprintf("Content for file %d", i)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}

		// Add the file
		if _, err := git.Exec(ctx, "add", fileName); err != nil {
			b.Fatal(err)
		}

		// Commit
		message := fmt.Sprintf("Commit %d", i)
		if _, err := git.Exec(ctx, "commit", "-m", message); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIsValidBranchNameLong benchmarks branch name validation with various lengths
func BenchmarkIsValidBranchNameLong(b *testing.B) {
	lengths := []int{10, 50, 100, 200}

	for _, length := range lengths {
		name := fmt.Sprintf("length-%d", length)
		branchName := "feature/" + string(make([]byte, length-8))
		for i := 8; i < length; i++ {
			branchName = branchName[:i] + "a"
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = isValidBranchName(branchName)
			}
		})
	}
}

// BenchmarkDetectWorktreeBranch benchmarks branch detection from worktree paths
func BenchmarkDetectWorktreeBranch(b *testing.B) {
	tmpDir := b.TempDir()
	setupBenchmarkRepo(b, tmpDir)
	git := New(tmpDir)
	ctx := context.Background()

	// Create a worktree
	branchName := "test-worktree-branch"
	worktreePath := filepath.Join(tmpDir, ".worktrees", branchName)
	if err := git.CreateBranch(ctx, branchName); err != nil {
		b.Fatal(err)
	}
	if err := git.AddWorktree(ctx, worktreePath, branchName); err != nil {
		b.Fatal(err)
	}

	// Change to worktree directory
	gitInWorktree := New(worktreePath)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gitInWorktree.CurrentBranch(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// setupBenchmarkRepo creates a git repository for benchmarking
func setupBenchmarkRepo(b *testing.B, tmpDir string) {
	b.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	// Configure git locally (not globally)
	cmd = exec.Command("git", "config", "user.name", "Benchmark User")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "benchmark@example.com")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	// Create initial file
	readmePath := filepath.Join(tmpDir, "README.md")
	require.NoError(b, os.WriteFile(readmePath, []byte("# Benchmark Repo"), 0644))

	// Initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	require.NoError(b, cmd.Run())
}
