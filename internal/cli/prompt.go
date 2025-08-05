package cli

import (
	"bufio"
	"fmt"
	"io"
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
	return PromptWithReader(message, options, os.Stdin)
}

// PromptWithReader displays a prompt and returns the selected option key using the provided reader
func PromptWithReader(message string, options []PromptOption, input io.Reader) (string, error) {
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
	reader := bufio.NewReader(input)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	userInput = strings.TrimSpace(strings.ToLower(userInput))

	// Use default if empty
	if userInput == "" && defaultKey != "" {
		return defaultKey, nil
	}

	// Validate input
	for _, opt := range options {
		if userInput == strings.ToLower(opt.Key) {
			return opt.Key, nil
		}
	}

	// Collect valid options for error message
	validKeys := make([]string, 0, len(options))
	for _, opt := range options {
		validKeys = append(validKeys, opt.Key)
	}

	return "", fmt.Errorf("invalid choice: %s (valid options: %s)", userInput, strings.Join(validKeys, ", "))
}

// ConfirmPrompt displays a yes/no prompt
func ConfirmPrompt(message string, defaultYes bool) bool {
	return ConfirmPromptWithReader(message, defaultYes, os.Stdin)
}

// ConfirmPromptWithReader displays a yes/no prompt using the provided reader
func ConfirmPromptWithReader(message string, defaultYes bool, input io.Reader) bool {
	suffix := " (y/N): "
	if defaultYes {
		suffix = " (Y/n): "
	}

	fmt.Print(message + suffix)

	reader := bufio.NewReader(input)
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(strings.ToLower(userInput))

	if userInput == "" {
		return defaultYes
	}

	return userInput == "y" || userInput == "yes"
}