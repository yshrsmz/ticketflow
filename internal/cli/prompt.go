package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptOption represents a choice in a prompt
type PromptOption struct {
	Key         string
	Description string
	IsDefault   bool
}

// Prompt displays a prompt and returns the selected option key
func Prompt(message string, options []PromptOption) (string, error) {
	fmt.Println(message)
	fmt.Println()

	// Display options
	var defaultKey string
	for _, opt := range options {
		if opt.IsDefault {
			defaultKey = opt.Key
			fmt.Printf("  [%s] %s (default)\n", opt.Key, opt.Description)
		} else {
			fmt.Printf("  [%s] %s\n", opt.Key, opt.Description)
		}
	}

	fmt.Print("\nYour choice: ")

	// Read input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	// Use default if empty
	if input == "" && defaultKey != "" {
		return defaultKey, nil
	}

	// Validate input
	for _, opt := range options {
		if input == strings.ToLower(opt.Key) {
			return opt.Key, nil
		}
	}

	return "", fmt.Errorf("invalid choice: %s", input)
}

// ConfirmPrompt displays a yes/no prompt
func ConfirmPrompt(message string, defaultYes bool) bool {
	suffix := " (y/N): "
	if defaultYes {
		suffix = " (Y/n): "
	}

	fmt.Print(message + suffix)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}