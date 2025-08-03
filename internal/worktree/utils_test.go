package worktree

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/mocks"
)

func TestGetPath(t *testing.T) {
	ctx := context.Background()
	projectRoot := "/test/project"
	ticketID := "test-ticket-123"

	tests := []struct {
		name         string
		setupMock    func(*mocks.MockGitClient)
		cfg          *config.Config
		expectedPath string
	}{
		{
			name: "returns actual worktree path when found",
			setupMock: func(m *mocks.MockGitClient) {
				m.On("FindWorktreeByBranch", ctx, ticketID).Return(
					&git.WorktreeInfo{
						Path:   "/actual/worktree/path",
						Branch: ticketID,
					}, nil)
			},
			cfg: &config.Config{
				Worktree: config.WorktreeConfig{
					BaseDir: ".worktrees",
				},
			},
			expectedPath: "/actual/worktree/path",
		},
		{
			name: "returns calculated path when worktree not found",
			setupMock: func(m *mocks.MockGitClient) {
				m.On("FindWorktreeByBranch", ctx, ticketID).Return(nil, nil)
			},
			cfg: &config.Config{
				Worktree: config.WorktreeConfig{
					BaseDir: ".worktrees",
				},
			},
			expectedPath: filepath.Join(projectRoot, ".worktrees", ticketID),
		},
		{
			name: "returns calculated path on error",
			setupMock: func(m *mocks.MockGitClient) {
				m.On("FindWorktreeByBranch", ctx, ticketID).Return(nil, assert.AnError)
			},
			cfg: &config.Config{
				Worktree: config.WorktreeConfig{
					BaseDir: "../custom-worktrees",
				},
			},
			expectedPath: filepath.Join(projectRoot, "../custom-worktrees", ticketID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit := new(mocks.MockGitClient)
			tt.setupMock(mockGit)

			path := GetPath(ctx, mockGit, tt.cfg, projectRoot, ticketID)
			assert.Equal(t, tt.expectedPath, path)

			mockGit.AssertExpectations(t)
		})
	}
}