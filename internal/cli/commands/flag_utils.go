package commands

import "flag"

// FlagResolver provides utilities for handling long/short form flag pairs.
// It follows the convention that short form flags take precedence when both are provided.
type FlagResolver struct{}

// StringFlag represents a string flag with long and short forms
type StringFlag struct {
	Long      string
	Short     string
	shortSet  bool
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