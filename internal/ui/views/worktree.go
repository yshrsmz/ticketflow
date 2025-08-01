package views

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yshrsmz/ticketflow/internal/config"
	"github.com/yshrsmz/ticketflow/internal/git"
	"github.com/yshrsmz/ticketflow/internal/ui/styles"
)

// WorktreeListModel represents the worktree list view
type WorktreeListModel struct {
	git        git.GitClient
	config     *config.Config
	worktrees  []git.WorktreeInfo
	cursor     int
	err        error
	shouldBack bool
	width      int
	height     int
	gitRoot    string // cached git root path
}

// NewWorktreeListModel creates a new worktree list model
func NewWorktreeListModel(g git.GitClient, cfg *config.Config) WorktreeListModel {
	root, _ := g.RootPath()
	return WorktreeListModel{
		git:     g,
		config:  cfg,
		gitRoot: root,
	}
}

// Init initializes the model
func (m WorktreeListModel) Init() tea.Cmd {
	return m.loadWorktrees()
}

// Update handles messages
func (m WorktreeListModel) Update(msg tea.Msg) (WorktreeListModel, tea.Cmd) {
	m.shouldBack = false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.shouldBack = true

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}

		case "g", "home":
			m.cursor = 0

		case "G", "end":
			if len(m.worktrees) > 0 {
				m.cursor = len(m.worktrees) - 1
			}

		case "r":
			return m, m.loadWorktrees()

		case "enter":
			// Could implement switching to worktree directory
			// For now, just show info
		}

	case worktreesLoadedMsg:
		m.worktrees = msg.worktrees
		m.err = msg.err
		if m.cursor >= len(m.worktrees) {
			m.cursor = len(m.worktrees) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case error:
		m.err = msg
	}

	return m, nil
}

// View renders the view
func (m WorktreeListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s\n", styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	if !m.config.Worktree.Enabled {
		return fmt.Sprintf("\n  %s\n\n  Enable worktrees in .ticketflow.yaml to use this feature.\n",
			styles.InfoStyle.Render("Worktrees are not enabled"))
	}

	if len(m.worktrees) == 0 {
		return fmt.Sprintf("\n  %s\n", styles.InfoStyle.Render("No worktrees found"))
	}

	var s strings.Builder

	// Title
	s.WriteString(styles.TitleStyle.Render("Git Worktrees"))
	s.WriteString("\n\n")

	// Calculate column widths
	branchWidth := 30
	pathWidth := m.width - branchWidth - 10
	if pathWidth < 20 {
		pathWidth = 20
	}

	// Header
	header := fmt.Sprintf("%-*s %s", branchWidth, "Branch", "Path")
	s.WriteString(styles.SubtitleStyle.Render(header))
	s.WriteString("\n")
	s.WriteString(strings.Repeat("─", m.width-4))
	s.WriteString("\n")

	// Worktree list
	for i, wt := range m.worktrees {
		branch := wt.Branch
		if branch == "" {
			branch = styles.MutedStyle.Render("(detached)")
		}

		// Make path relative if possible
		path := wt.Path
		if rel, err := filepath.Rel(m.gitRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
			path = rel
		}

		// Highlight main worktree
		if wt.Branch == "" || strings.Contains(path, m.gitRoot) && !strings.Contains(path, m.config.Worktree.BaseDir) {
			branch = styles.SuccessStyle.Render("main")
		}

		row := fmt.Sprintf("%-*s %s",
			branchWidth, truncate(branch, branchWidth),
			truncate(path, pathWidth))

		// Apply cursor styling
		if i == m.cursor {
			row = styles.SelectedItemStyle.Render(row)
		} else {
			row = styles.ItemStyle.Render(row)
		}

		s.WriteString(row)
		s.WriteString("\n")
	}

	// Selected worktree details
	if m.cursor >= 0 && m.cursor < len(m.worktrees) {
		selected := m.worktrees[m.cursor]
		s.WriteString("\n")

		detailBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.GetPriorityStyle(3).GetForeground()).
			Padding(0, 1).
			Width(m.width - 4)

		var details strings.Builder
		details.WriteString(fmt.Sprintf("Path: %s\n", selected.Path))
		if selected.Branch != "" {
			details.WriteString(fmt.Sprintf("Branch: %s\n", selected.Branch))
		}
		if selected.HEAD != "" {
			details.WriteString(fmt.Sprintf("HEAD: %s", selected.HEAD[:8]))
		}

		s.WriteString(detailBox.Render(details.String()))
	}

	// Help
	s.WriteString("\n\n")
	s.WriteString(styles.HelpStyle.Render("Press q to go back • r to refresh"))

	return s.String()
}

// SetSize sets the view size
func (m *WorktreeListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShouldGoBack returns whether the view should go back
func (m WorktreeListModel) ShouldGoBack() bool {
	return m.shouldBack
}

// worktreesLoadedMsg is sent when worktrees are loaded
type worktreesLoadedMsg struct {
	worktrees []git.WorktreeInfo
	err       error
}

// loadWorktrees loads the worktree list
func (m WorktreeListModel) loadWorktrees() tea.Cmd {
	return func() tea.Msg {
		worktrees, err := m.git.ListWorktrees()
		return worktreesLoadedMsg{
			worktrees: worktrees,
			err:       err,
		}
	}
}
