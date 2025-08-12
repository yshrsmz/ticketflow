package command

import (
	"fmt"
	"sync"
)

// registry is the default implementation of Registry
type registry struct {
	mu       sync.RWMutex
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() Registry {
	return &registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *registry) Register(cmd Command) error {
	if cmd == nil {
		return fmt.Errorf("cannot register nil command")
	}

	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("cannot register command with empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %q already registered", name)
	}

	r.commands[name] = cmd
	return nil
}

// Get retrieves a command by name
func (r *registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	return cmd, exists
}

// List returns all registered commands
func (r *registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}
