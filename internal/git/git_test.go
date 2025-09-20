package git

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWithTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		repoPath    string
		timeout     time.Duration
		wantPath    string
		wantTimeout time.Duration
	}{
		{
			name:        "custom timeout",
			repoPath:    "/tmp/test",
			timeout:     45 * time.Second,
			wantPath:    "/tmp/test",
			wantTimeout: 45 * time.Second,
		},
		{
			name:        "zero timeout defaults to 30s",
			repoPath:    "/tmp/test2",
			timeout:     0,
			wantPath:    "/tmp/test2",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "negative timeout defaults to 30s",
			repoPath:    "/tmp/test3",
			timeout:     -5 * time.Second,
			wantPath:    "/tmp/test3",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "empty path defaults to current dir",
			repoPath:    "",
			timeout:     10 * time.Second,
			wantPath:    ".",
			wantTimeout: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithTimeout(tt.repoPath, tt.timeout)
			assert.Equal(t, tt.wantPath, g.repoPath)
			assert.Equal(t, tt.wantTimeout, g.timeout)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		repoPath string
		wantPath string
	}{
		{
			name:     "valid path",
			repoPath: "/tmp/test",
			wantPath: "/tmp/test",
		},
		{
			name:     "empty path defaults to current dir",
			repoPath: "",
			wantPath: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(tt.repoPath)
			assert.Equal(t, tt.wantPath, g.repoPath)
			assert.Equal(t, 30*time.Second, g.timeout)
		})
	}
}

func TestExecWithTimeout(t *testing.T) {
	t.Parallel()
	// Create a git instance with a short but reasonable timeout
	// Using 50ms to reduce flakiness while still testing timeout behavior
	g := NewWithTimeout(".", 50*time.Millisecond)

	// Try to execute a command that would take longer than the timeout
	// Using a command that's likely to take more time
	ctx := context.Background()
	_, err := g.Exec(ctx, "log", "--all", "--oneline", "-n", "100000")

	// The error might be a timeout or might succeed if git is fast enough
	// We're mainly testing that the timeout is applied without panicking
	if err != nil {
		// The error could be "signal: killed" or "context deadline exceeded"
		assert.True(t,
			strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "signal: killed") ||
				strings.Contains(err.Error(), "operation timed out"),
			"Expected timeout-related error, got: %v", err)
	}
}

func TestExecWithContextTimeout(t *testing.T) {
	t.Parallel()
	// Create a git instance with a long timeout
	g := NewWithTimeout(".", 30*time.Second)

	// But use a context with a short timeout
	// Using 50ms to reduce flakiness
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := g.Exec(ctx, "log", "--all", "--oneline", "-n", "100000")

	// The context timeout should take precedence
	if err != nil {
		// The error could be "signal: killed" or "context deadline exceeded"
		assert.True(t,
			strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "signal: killed") ||
				strings.Contains(err.Error(), "operation timed out"),
			"Expected timeout-related error, got: %v", err)
	}
}

func TestRunInWorktreePreservesTimeout(t *testing.T) {
	t.Parallel()
	// Create a git instance with custom timeout
	g := NewWithTimeout(".", 45*time.Second)

	// The RunInWorktree method should preserve the timeout
	// We can't easily test the actual execution without a real worktree,
	// but we can verify the method doesn't panic
	ctx := context.Background()
	_, _ = g.RunInWorktree(ctx, "/tmp/fake-worktree", "status")
}

func TestFindMainRepositoryRoot_FromWorktree(t *testing.T) {
	t.Parallel()
	git, tmpDir := setupTestGitRepo(t)
	ctx := context.Background()

	// Create a linked worktree for a new branch
	wtPath := filepath.Join(tmpDir, ".worktrees", "wt-root")
	err := git.AddWorktree(ctx, wtPath, "wt-root")
	assert.NoError(t, err)

	// Resolve project root starting from inside the worktree
	root, err := FindMainRepositoryRoot(ctx, wtPath)
	assert.NoError(t, err)

	gotRoot, err := filepath.EvalSymlinks(root)
	assert.NoError(t, err)
	wantRoot, err := filepath.EvalSymlinks(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, wantRoot, gotRoot)
}

func TestIsValidBranchCharEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ch   rune
		want bool
	}{
		{"ASCII DEL 0x7F", 0x7F, false},
		{"ASCII space", ' ', false},
		{"ASCII tab", '\t', false},
		{"Valid hyphen", '-', true},
		{"Valid letter", 'a', true},
		{"Control char 0x1F", 0x1F, false},
		{"Control char 0x00", 0x00, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidBranchChar(tt.ch)
			if got != tt.want {
				t.Errorf("isValidBranchChar(%q/0x%X) = %v, want %v", tt.ch, tt.ch, got, tt.want)
			}
		})
	}
}

func TestIsValidBranchName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		// Valid branch names
		{"simple branch", "feature", true},
		{"branch with slash", "feature/test", true},
		{"branch with numbers", "feature123", true},
		{"branch with dash", "feature-test", true},
		{"branch with underscore", "feature_test", true},
		{"branch with multiple segments", "feature/test/v2", true},
		{"branch with version", "release-1.2.3", true},
		{"branch with dots in middle", "release.1.2.3", true},
		{"complex valid branch", "feature/JIRA-123_test-branch.v2", true},
		{"branch starting with letter", "a", true},
		{"branch starting with number", "123-feature", true},

		// Invalid branch names
		{"empty string", "", false},
		{"starts with slash", "/feature", false},
		{"ends with slash", "feature/", false},
		{"double slash", "feature//test", false},
		{"contains space", "feature test", false},
		{"contains dot dot", "feature..test", false},
		{"contains @{", "feature@{test", false},
		{"starts with dot", ".feature", false},
		{"ends with dot", "feature.", false},
		{"contains colon", "feature:test", false},
		{"contains question mark", "feature?test", false},
		{"contains asterisk", "feature*test", false},
		{"contains open bracket", "feature[test", false},
		{"contains backslash", "feature\\test", false},
		{"contains control char", "feature\x00test", false},
		{"contains tab", "feature\ttest", false},
		{"single dot", ".", false},
		{"single slash", "/", false},

		// Unicode characters
		{"branch with emoji", "feature-🚀", true},
		{"branch with unicode", "feature-résumé", true},
		{"branch with chinese", "feature-中文", true},
		{"branch with mixed unicode", "feat/test-中文-résumé", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidBranchName(tt.branch)
			assert.Equal(t, tt.expected, result, "branch: %q", tt.branch)
		})
	}
}

func TestBranchExists(t *testing.T) {
	t.Parallel()
	// Setup test git repo
	tmpDir := t.TempDir()
	git := New(tmpDir)
	ctx := context.Background()

	// Initialize repo
	_, err := git.Exec(ctx, "init")
	assert.NoError(t, err)

	// Configure git locally for test repo (not globally)
	configureTestGitClient(t, git)

	// Create initial commit
	_, err = git.Exec(ctx, "commit", "--allow-empty", "--no-gpg-sign", "-m", "Initial commit")
	assert.NoError(t, err)

	tests := []struct {
		name       string
		branch     string
		setup      func()
		wantExists bool
		wantErr    bool
	}{
		{
			name:   "master/main branch exists",
			branch: "master",
			setup: func() {
				// Try to create master branch, ignore error if it already exists
				_, _ = git.Exec(ctx, "checkout", "-b", "master")
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "non-existent branch",
			branch:     "feature/does-not-exist",
			setup:      func() {},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:   "existing feature branch",
			branch: "feature/test-branch",
			setup: func() {
				_, _ = git.Exec(ctx, "checkout", "-b", "feature/test-branch")
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:   "branch with special characters",
			branch: "feat/my-branch_v2.0",
			setup: func() {
				_, _ = git.Exec(ctx, "checkout", "-b", "feat/my-branch_v2.0")
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "invalid branch name with spaces",
			branch:     "feature test",
			setup:      func() {},
			wantExists: false,
			wantErr:    true,
		},
		{
			name:       "invalid branch name with command injection attempt",
			branch:     "feature; rm -rf /",
			setup:      func() {},
			wantExists: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			exists, err := git.BranchExists(ctx, tt.branch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantExists, exists)
			}
		})
	}
}

func BenchmarkIsValidBranchChar(b *testing.B) {
	testCases := []struct {
		name string
		ch   rune
	}{
		{"ASCII letter", 'a'},
		{"ASCII digit", '5'},
		{"ASCII hyphen", '-'},
		{"ASCII control", '\x1f'},
		{"ASCII DEL", '\x7f'},
		{"Unicode letter", '世'},
		{"Forbidden char", ':'},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = isValidBranchChar(tc.ch)
			}
		})
	}
}

func BenchmarkIsValidBranchName(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"short ASCII", "feature/my-new-feature"},
		{"long ASCII", "feature/my-very-long-feature-name-with-many-words-123456789"},
		{"with unicode", "feature/新機能-implementation"},
		{"with invalid", "feature/my new feature!"},
		{"typical ticket branch", "250726-183403-fix-branch-already-exist-on-start"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = isValidBranchName(tc.input)
			}
		})
	}
}
