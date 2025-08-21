// Package commands provides CLI command implementations and utilities.
// This file contains flag handling utilities that manage long/short form flag pairs
// with proper precedence handling (short form takes precedence when explicitly set).
//
// Thread-safety note: These types are designed for single-threaded flag parsing
// at program startup. They are not safe for concurrent modification during parsing.
package commands

import "flag"

// StringFlag represents a string flag with long and short forms.
// The shortSet field tracks whether the short form was explicitly set via command line,
// allowing proper precedence handling where short form overrides long form.
type StringFlag struct {
	Long     string
	Short    string
	shortSet bool // tracks if short form was explicitly set (not thread-safe)
}

// BoolFlag represents a bool flag with long and short forms
type BoolFlag struct {
	Long  bool
	Short bool
}

// RegisterString registers both long and short form string flags.
// The short form uses a custom handler to detect if it was explicitly set.
// At least one of longName or shortName must be provided.
func RegisterString(fs *flag.FlagSet, sf *StringFlag, longName, shortName, defaultValue, usage string) {
	if longName == "" && shortName == "" {
		panic("RegisterString: at least one of longName or shortName must be provided")
	}

	sf.Long = defaultValue

	// Register long form with default if provided
	if longName != "" {
		fs.StringVar(&sf.Long, longName, defaultValue, usage)
	}

	// Register short form with custom handler to track if it was set
	if shortName != "" {
		fs.Func(shortName, usage+" (short form)", func(value string) error {
			sf.Short = value
			sf.shortSet = true
			return nil
		})
	}
}

// RegisterBool registers both long and short form bool flags.
// At least one of longName or shortName must be provided.
func RegisterBool(fs *flag.FlagSet, bf *BoolFlag, longName, shortName string, usage string) {
	if longName == "" && shortName == "" {
		panic("RegisterBool: at least one of longName or shortName must be provided")
	}

	if longName != "" {
		fs.BoolVar(&bf.Long, longName, false, usage)
	}
	if shortName != "" {
		fs.BoolVar(&bf.Short, shortName, false, usage+" (short form)")
	}
}

// Value returns the resolved string value (short takes precedence if set)
func (sf *StringFlag) Value() string {
	if sf.shortSet {
		return sf.Short
	}
	return sf.Long
}

// Value returns the resolved bool value (OR of both flags)
func (bf *BoolFlag) Value() bool {
	return bf.Long || bf.Short
}

// Removed MergeFlags struct - it was an early design that's no longer needed.
// The individual StringFlag and BoolFlag types with their Value() methods
// provide a cleaner, more composable solution.
