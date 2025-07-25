# TicketFlow - Implementation Examples

## 1. チケットモデルの実装例

### internal/ticket/ticket.go

```go
package ticket

import (
    "bytes"
    "fmt"
    "strings"
    "time"
    
    "gopkg.in/yaml.v3"
)

// Status represents the ticket status
type Status string

const (
    StatusTodo  Status = "todo"
    StatusDoing Status = "doing"
    StatusDone  Status = "done"
)

// Ticket represents a ticket with metadata and content
type Ticket struct {
    // Metadata (YAML frontmatter)
    Priority    int        `yaml:"priority"`
    Description string     `yaml:"description"`
    CreatedAt   time.Time  `yaml:"created_at"`
    StartedAt   *time.Time `yaml:"started_at"`
    ClosedAt    *time.Time `yaml:"closed_at"`
    Related     []string   `yaml:"related,omitempty"`
    
    // Computed fields
    ID      string `yaml:"-"`
    Slug    string `yaml:"-"`
    Path    string `yaml:"-"`
    Content string `yaml:"-"`
}

// Status returns the current status of the ticket
func (t *Ticket) Status() Status {
    if t.ClosedAt != nil {
        return StatusDone
    }
    if t.StartedAt != nil {
        return StatusDoing
    }
    return StatusTodo
}

// HasWorktree checks if the ticket has an associated worktree
func (t *Ticket) HasWorktree() bool {
    return t.Status() == StatusDoing
}

// Parse parses a ticket file content
func Parse(content []byte) (*Ticket, error) {
    parts := bytes.SplitN(content, []byte("---\n"), 3)
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid ticket format: missing frontmatter")
    }
    
    // Parse YAML frontmatter
    var ticket Ticket
    if err := yaml.Unmarshal(parts[1], &ticket); err != nil {
        return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
    }
    
    // Set content (remove leading newline if present)
    ticket.Content = strings.TrimPrefix(string(parts[2]), "\n")
    
    return &ticket, nil
}

// ToBytes converts the ticket to file content
func (t *Ticket) ToBytes() ([]byte, error) {
    var buf bytes.Buffer
    
    // Write frontmatter
    buf.WriteString("---\n")
    
    encoder := yaml.NewEncoder(&buf)
    encoder.SetIndent(0)
    if err := encoder.Encode(t); err != nil {
        return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
    }
    
    buf.WriteString("---\n\n")
    buf.WriteString(t.Content)
    
    return buf.Bytes(), nil
}

// GenerateID generates a ticket ID with current timestamp
func GenerateID(slug string) string {
    now := time.Now()
    return fmt.Sprintf("%s-%s",
        now.Format("060102-150405"),
        slug,
    )
}

// ParseID extracts timestamp and slug from ticket ID
func ParseID(id string) (time.Time, string, error) {
    parts := strings.SplitN(id, "-", 3)
    if len(parts) < 3 {
        return time.Time{}, "", fmt.Errorf("invalid ticket ID format")
    }
    
    // Parse timestamp
    timestamp, err := time.Parse("060102-150405", 
        fmt.Sprintf("%s-%s", parts[0], parts[1]))
    if err != nil {
        return time.Time{}, "", fmt.Errorf("invalid timestamp: %w", err)
    }
    
    slug := parts[2]
    return timestamp, slug, nil
}
```

## 2. Git操作の実装例

### internal/git/git.go

```go
package git

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
)

// Git provides git operations
type Git struct {
    repoPath string
}

// New creates a new Git instance
func New(repoPath string) *Git {
    return &Git{repoPath: repoPath}
}

// exec executes a git command
func (g *Git) exec(args ...string) (string, error) {
    cmd := exec.Command("git", args...)
    cmd.Dir = g.repoPath
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("git %s failed: %w\n%s", 
            strings.Join(args, " "), err, stderr.String())
    }
    
    return strings.TrimSpace(stdout.String()), nil
}

// CurrentBranch returns the current branch name
func (g *Git) CurrentBranch() (string, error) {
    return g.exec("rev-parse", "--abbrev-ref", "HEAD")
}

// CreateBranch creates and checks out a new branch
func (g *Git) CreateBranch(name string) error {
    _, err := g.exec("checkout", "-b", name)
    return err
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *Git) HasUncommittedChanges() (bool, error) {
    output, err := g.exec("status", "--porcelain")
    if err != nil {
        return false, err
    }
    return output != "", nil
}

// Add stages files
func (g *Git) Add(files ...string) error {
    args := append([]string{"add"}, files...)
    _, err := g.exec(args...)
    return err
}

// Commit creates a commit
func (g *Git) Commit(message string) error {
    _, err := g.exec("commit", "-m", message)
    return err
}

// Checkout switches to a branch
func (g *Git) Checkout(branch string) error {
    _, err := g.exec("checkout", branch)
    return err
}

// MergeSquash performs a squash merge
func (g *Git) MergeSquash(branch string) error {
    _, err := g.exec("merge", "--squash", branch)
    return err
}

// Push pushes a branch to remote
func (g *Git) Push(remote, branch string, setUpstream bool) error {
    args := []string{"push"}
    if setUpstream {
        args = append(args, "-u")
    }
    args = append(args, remote, branch)
    _, err := g.exec(args...)
    return err
}

// IsGitRepo checks if the path is a git repository
func IsGitRepo(path string) bool {
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    cmd.Dir = path
    return cmd.Run() == nil
}
```

### internal/git/worktree.go

```go
package git

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// WorktreeInfo represents worktree information
type WorktreeInfo struct {
    Path   string
    Branch string
    HEAD   string
}

// ListWorktrees lists all worktrees
func (g *Git) ListWorktrees() ([]WorktreeInfo, error) {
    output, err := g.exec("worktree", "list", "--porcelain")
    if err != nil {
        return nil, err
    }
    
    var worktrees []WorktreeInfo
    lines := strings.Split(output, "\n")
    
    var current WorktreeInfo
    for _, line := range lines {
        if line == "" {
            if current.Path != "" {
                worktrees = append(worktrees, current)
                current = WorktreeInfo{}
            }
            continue
        }
        
        parts := strings.SplitN(line, " ", 2)
        if len(parts) != 2 {
            continue
        }
        
        switch parts[0] {
        case "worktree":
            current.Path = parts[1]
        case "HEAD":
            current.HEAD = parts[1]
        case "branch":
            current.Branch = strings.TrimPrefix(parts[1], "refs/heads/")
        }
    }
    
    if current.Path != "" {
        worktrees = append(worktrees, current)
    }
    
    return worktrees, nil
}

// AddWorktree creates a new worktree
func (g *Git) AddWorktree(path, branch string) error {
    // Create parent directory if needed
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return fmt.Errorf("failed to create worktree directory: %w", err)
    }
    
    _, err := g.exec("worktree", "add", path, "-b", branch)
    return err
}

// RemoveWorktree removes a worktree
func (g *Git) RemoveWorktree(path string) error {
    _, err := g.exec("worktree", "remove", path, "--force")
    return err
}

// PruneWorktrees removes worktree information for deleted directories
func (g *Git) PruneWorktrees() error {
    _, err := g.exec("worktree", "prune")
    return err
}
```

## 3. CLIコマンドの実装例

### internal/cli/commands.go

```go
package cli

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/yshrsmz/ticketflow/internal/config"
    "github.com/yshrsmz/ticketflow/internal/git"
    "github.com/yshrsmz/ticketflow/internal/ticket"
)

// App represents the CLI application
type App struct {
    Config  *config.Config
    Git     *git.Git
    Manager *ticket.Manager
}

// NewApp creates a new CLI application
func NewApp() (*App, error) {
    // Find project root (with .git directory)
    projectRoot, err := findProjectRoot()
    if err != nil {
        return nil, NewError(ErrNotGitRepo, "Not in a git repository", "", 
            []string{
                "Navigate to your project root directory",
                "Initialize a new git repository with 'git init'",
            })
    }
    
    // Load config
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, NewError(ErrConfigNotFound, "Ticket system not initialized", "", 
            []string{
                "Run 'ticketflow init' to initialize",
                "Navigate to the project root directory",
            })
    }
    
    gitClient := git.New(projectRoot)
    manager := ticket.NewManager(cfg, gitClient)
    
    return &App{
        Config:  cfg,
        Git:     gitClient,
        Manager: manager,
    }, nil
}

// Init initializes the ticket system
func (app *App) Init() error {
    projectRoot, err := findProjectRoot()
    if err != nil {
        return NewError(ErrNotGitRepo, "Not in a git repository", "", nil)
    }
    
    // Create default config
    cfg := config.Default()
    configPath := filepath.Join(projectRoot, ".ticketflow.yaml")
    
    // Check if already exists
    if _, err := os.Stat(configPath); err == nil {
        fmt.Println("Ticket system already initialized")
        return nil
    }
    
    // Save config
    if err := cfg.Save(configPath); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    
    // Create directories
    ticketsDir := filepath.Join(projectRoot, cfg.Tickets.Dir)
    if err := os.MkdirAll(ticketsDir, 0755); err != nil {
        return fmt.Errorf("failed to create tickets directory: %w", err)
    }
    
    // Update .gitignore
    gitignorePath := filepath.Join(projectRoot, ".gitignore")
    if err := updateGitignore(gitignorePath); err != nil {
        return fmt.Errorf("failed to update .gitignore: %w", err)
    }
    
    fmt.Println("Initialized ticket system successfully")
    fmt.Printf("Configuration saved to: %s\n", configPath)
    fmt.Printf("Tickets directory: %s\n", ticketsDir)
    
    return nil
}

// NewTicket creates a new ticket
func (app *App) NewTicket(slug string, format OutputFormat) error {
    // Validate slug
    if !isValidSlug(slug) {
        return NewError(ErrTicketInvalid, "Invalid slug format", 
            fmt.Sprintf("Slug '%s' contains invalid characters", slug),
            []string{
                "Use only lowercase letters (a-z)",
                "Use only numbers (0-9)",
                "Use only hyphens (-) for separation",
            })
    }
    
    // Create ticket
    ticket, err := app.Manager.Create(slug)
    if err != nil {
        return err
    }
    
    if format == FormatJSON {
        return outputJSON(map[string]interface{}{
            "ticket": map[string]interface{}{
                "id":   ticket.ID,
                "path": ticket.Path,
            },
        })
    }
    
    fmt.Printf("Created ticket file: %s\n", ticket.Path)
    fmt.Println("Please edit the file to add title, description and details.")
    
    return nil
}

// ListTickets lists tickets
func (app *App) ListTickets(status ticket.Status, count int, format OutputFormat) error {
    tickets, err := app.Manager.List(status)
    if err != nil {
        return err
    }
    
    // Limit count
    if count > 0 && len(tickets) > count {
        tickets = tickets[:count]
    }
    
    if format == FormatJSON {
        return app.outputTicketListJSON(tickets)
    }
    
    return app.outputTicketListText(tickets)
}

// Helper functions

func findProjectRoot() (string, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return "", err
    }
    
    dir := cwd
    for {
        if git.IsGitRepo(dir) {
            return dir, nil
        }
        
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    
    return "", fmt.Errorf("not in a git repository")
}

func isValidSlug(slug string) bool {
    if slug == "" {
        return false
    }
    
    for _, r := range slug {
        if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
            return false
        }
    }
    
    return true
}

func outputJSON(data interface{}) error {
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}
```

## 4. TUIの実装例

### internal/ui/app.go

```go
package ui

import (
    "fmt"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yshrsmz/ticketflow/internal/ticket"
)

// ViewType represents the current view
type ViewType int

const (
    ViewTicketList ViewType = iota
    ViewTicketDetail
    ViewNewTicket
    ViewWorktreeList
    ViewHelp
)

// Model represents the application state
type Model struct {
    view         ViewType
    ticketList   ticketListModel
    ticketDetail ticketDetailModel
    newTicket    newTicketModel
    worktreeList worktreeListModel
    help         helpModel
    
    manager *ticket.Manager
    width   int
    height  int
    err     error
}

// New creates a new TUI application
func New(manager *ticket.Manager) Model {
    return Model{
        view:         ViewTicketList,
        ticketList:   newTicketListModel(),
        ticketDetail: newTicketDetailModel(),
        newTicket:    newNewTicketModel(),
        worktreeList: newWorktreeListModel(),
        help:         newHelpModel(),
        manager:      manager,
    }
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.ticketList.init(),
        tea.SetWindowTitle("TicketFlow"),
    )
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global keybindings
        switch msg.String() {
        case "ctrl+c", "q":
            if m.view == ViewTicketList {
                return m, tea.Quit
            }
        case "?":
            m.view = ViewHelp
            return m, nil
        }
        
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        
    case error:
        m.err = msg
        return m, nil
    }
    
    // Delegate to current view
    var cmd tea.Cmd
    switch m.view {
    case ViewTicketList:
        m.ticketList, cmd = m.ticketList.Update(msg)
        
        // Handle view transitions
        if m.ticketList.shouldShowDetail {
            m.view = ViewTicketDetail
            m.ticketDetail = m.ticketDetail.setTicket(m.ticketList.selectedTicket())
        } else if m.ticketList.shouldCreateNew {
            m.view = ViewNewTicket
        }
        
    case ViewTicketDetail:
        m.ticketDetail, cmd = m.ticketDetail.Update(msg)
        if m.ticketDetail.shouldGoBack {
            m.view = ViewTicketList
        }
        
    case ViewNewTicket:
        m.newTicket, cmd = m.newTicket.Update(msg)
        if m.newTicket.shouldGoBack {
            m.view = ViewTicketList
            if m.newTicket.created {
                // Refresh ticket list
                cmd = m.ticketList.refresh()
            }
        }
        
    case ViewHelp:
        m.help, cmd = m.help.Update(msg)
        if m.help.shouldClose {
            m.view = ViewTicketList
        }
    }
    
    return m, cmd
}

// View renders the UI
func (m Model) View() string {
    if m.err != nil {
        return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
    }
    
    switch m.view {
    case ViewTicketList:
        return m.ticketList.View()
    case ViewTicketDetail:
        return m.ticketDetail.View()
    case ViewNewTicket:
        return m.newTicket.View()
    case ViewWorktreeList:
        return m.worktreeList.View()
    case ViewHelp:
        return m.help.View()
    default:
        return "Unknown view"
    }
}
```

## 5. main.goの実装例

### cmd/ticketflow/main.go

```go
package main

import (
    "flag"
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yshrsmz/ticketflow/internal/cli"
    "github.com/yshrsmz/ticketflow/internal/ui"
)

func main() {
    // No arguments = TUI mode
    if len(os.Args) == 1 {
        runTUI()
        return
    }
    
    // CLI mode
    if err := runCLI(); err != nil {
        cli.HandleError(err)
        os.Exit(1)
    }
}

func runTUI() {
    app, err := cli.NewApp()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    program := tea.NewProgram(ui.New(app.Manager))
    if _, err := program.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func runCLI() error {
    // Define subcommands
    initCmd := flag.NewFlagSet("init", flag.ExitOnError)
    
    newCmd := flag.NewFlagSet("new", flag.ExitOnError)
    
    listCmd := flag.NewFlagSet("list", flag.ExitOnError)
    listStatus := listCmd.String("status", "", "Filter by status (todo|doing|done)")
    listCount := listCmd.Int("count", 20, "Maximum number of tickets to show")
    listFormat := listCmd.String("format", "text", "Output format (text|json)")
    
    startCmd := flag.NewFlagSet("start", flag.ExitOnError)
    startNoPush := startCmd.Bool("no-push", false, "Don't push branch to remote")
    
    closeCmd := flag.NewFlagSet("close", flag.ExitOnError)
    closeNoPush := closeCmd.Bool("no-push", false, "Don't push to remote")
    closeForce := closeCmd.Bool("force", false, "Force close with uncommitted changes")
    closeForceShort := closeCmd.Bool("f", false, "Force close (short form)")
    
    // Parse command
    if len(os.Args) < 2 {
        printUsage()
        return nil
    }
    
    switch os.Args[1] {
    case "init":
        initCmd.Parse(os.Args[2:])
        return handleInit()
        
    case "new":
        newCmd.Parse(os.Args[2:])
        if newCmd.NArg() < 1 {
            return fmt.Errorf("missing slug argument")
        }
        return handleNew(newCmd.Arg(0), *listFormat)
        
    case "list":
        listCmd.Parse(os.Args[2:])
        return handleList(*listStatus, *listCount, *listFormat)
        
    case "start":
        startCmd.Parse(os.Args[2:])
        if startCmd.NArg() < 1 {
            return fmt.Errorf("missing ticket argument")
        }
        return handleStart(startCmd.Arg(0), *startNoPush)
        
    case "close":
        closeCmd.Parse(os.Args[2:])
        force := *closeForce || *closeForceShort
        return handleClose(*closeNoPush, force)
        
    case "help", "-h", "--help":
        printUsage()
        return nil
        
    default:
        return fmt.Errorf("unknown command: %s", os.Args[1])
    }
}

func handleInit() error {
    // Special case: init doesn't require existing config
    return cli.InitCommand()
}

func handleNew(slug, format string) error {
    app, err := cli.NewApp()
    if err != nil {
        return err
    }
    
    outputFormat := cli.ParseOutputFormat(format)
    return app.NewTicket(slug, outputFormat)
}

func handleList(status string, count int, format string) error {
    app, err := cli.NewApp()
    if err != nil {
        return err
    }
    
    var ticketStatus ticket.Status
    if status != "" {
        ticketStatus = ticket.Status(status)
        if !isValidStatus(ticketStatus) {
            return fmt.Errorf("invalid status: %s", status)
        }
    }
    
    outputFormat := cli.ParseOutputFormat(format)
    return app.ListTickets(ticketStatus, count, outputFormat)
}

// ... other handlers ...

func printUsage() {
    fmt.Println(`Ticket Management System for Coding Agents

USAGE:
  ticketflow                          Start TUI (interactive mode)
  ticketflow init                     Initialize ticket system
  ticketflow new <slug>               Create new ticket
  ticketflow list [options]           List tickets
  ticketflow start <ticket> [options] Start working on ticket
  ticketflow close [options]          Complete current ticket
  ticketflow restore                  Restore current-ticket link
  ticketflow worktree <command>       Manage worktrees
  ticketflow help                     Show this help

Use 'ticketflow <command> -h' for command-specific help.`)
}
```

これらのドキュメントとコード例を参考に、Claude Codeで実装を進めてください。Phase 1から順番に実装し、各フェーズが完了したら次に進むようにしてください。
