package command

import (
	"fmt"
	"sync"
)

// registry is the default implementation of Registry
type registry struct {
	mu       sync.RWMutex
	commands map[string]Command
	aliases  map[string]string // maps alias to command name
}

// NewRegistry creates a new command registry
func NewRegistry() Registry {
	return &registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
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

	// Register aliases
	for _, alias := range cmd.Aliases() {
		if existingCmd, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %q already registered for command %q", alias, existingCmd)
		}
		r.aliases[alias] = name
	}

	return nil
}

// Get retrieves a command by name or alias
func (r *registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if it's a direct command name
	if cmd, exists := r.commands[name]; exists {
		return cmd, true
	}

	// Check if it's an alias
	if cmdName, exists := r.aliases[name]; exists {
		return r.commands[cmdName], true
	}

	return nil, false
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
