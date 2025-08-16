package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptOption represents a choice in a prompt
type PromptOption struct {
	Key         string
	Description string
	IsDefault   bool
}

// IsInteractive checks if the current environment is interactive
func IsInteractive() bool {
	// Check for explicit non-interactive environment variable
	if os.Getenv("TICKETFLOW_NON_INTERACTIVE") == "true" {
		return false
	}

	// Check common CI environment variables
	ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "JENKINS_URL"}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return false
		}
	}

	// Check if stdin is a terminal
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Prompt displays a prompt and returns the selected option key
func Prompt(message string, options []PromptOption) (string, error) {
	// Use a null status writer for backward compatibility
	return PromptWithStatus(message, options, &nullStatusWriter{})
}

// PromptWithStatus displays a prompt and returns the selected option key with status output
func PromptWithStatus(message string, options []PromptOption, status StatusWriter) (string, error) {
	// Ensure we have a status writer (defensive programming)
	if status == nil {
		status = &nullStatusWriter{}
	}

	// In non-interactive mode, automatically use the default option
	if !IsInteractive() {
		for _, opt := range options {
			if opt.IsDefault {
				status.Printf("Non-interactive mode detected. Using default option: %s\n", opt.Key)
				return opt.Key, nil
			}
		}
		// If no default option is set, return an error
		return "", fmt.Errorf("non-interactive mode detected and no default option available")
	}

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
	// Use a null status writer for backward compatibility
	return ConfirmPromptWithStatus(message, defaultYes, &nullStatusWriter{})
}

// ConfirmPromptWithStatus displays a yes/no prompt with status output
func ConfirmPromptWithStatus(message string, defaultYes bool, status StatusWriter) bool {
	// Ensure we have a status writer (defensive programming)
	if status == nil {
		status = &nullStatusWriter{}
	}

	// In non-interactive mode, automatically use the default
	if !IsInteractive() {
		action := "No"
		if defaultYes {
			action = "Yes"
		}
		status.Printf("Non-interactive mode detected. Using default: %s\n", action)
		return defaultYes
	}

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
