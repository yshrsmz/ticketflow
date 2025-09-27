// Package commands provides CLI command implementations and utilities.
// This file contains flag type definitions that manage long/short form flag pairs
// with proper precedence handling (short form takes precedence when explicitly set).
//
// Thread-safety note: These types are designed for single-threaded flag parsing
// at program startup. They are not safe for concurrent modification during parsing.
package commands

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"reflect"
)

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
		panic(fmt.Sprintf("RegisterString: at least one of longName or shortName must be provided (got longName=%q, shortName=%q)", longName, shortName))
	}

	sf.Long = defaultValue

	// TODO(Phase 2): Remove this reflection-based workaround
	// This is intentional technical debt from Phase 1 of the pflag migration.
	// Phase 1 (ticket: 250926-165945-phase1-pflag-basic-import-migration) replaced imports only.
	// Phase 2 (ticket: 250926-170130-phase2-pflag-proper-registration) will:
	//   - Remove all reflection-based code
	//   - Use pflag's StringVarP directly
	//   - Properly handle flag precedence with pflag's native behavior
	// See docs/pflag-migration-phases.md for full migration plan
	//
	// Phase 1 pflag compatibility: Use pflag's StringVarP when both are provided
	// This is a temporary fix for Phase 1 - will be properly refactored in Phase 2
	if longName != "" && shortName != "" {
		// StringVarP is a pflag-specific method that registers both long and short forms
		// We use reflection to call it if available (when using pflag)
		if method := reflect.ValueOf(fs).MethodByName("StringVarP"); method.IsValid() {
			// Call StringVarP(p *string, name, shorthand string, value string, usage string)
			method.Call([]reflect.Value{
				reflect.ValueOf(&sf.Long),
				reflect.ValueOf(longName),
				reflect.ValueOf(shortName),
				reflect.ValueOf(defaultValue),
				reflect.ValueOf(usage),
			})
			// Note: With pflag, both short and long forms write to the same variable (sf.Long)
			// The "last flag wins" behavior is pflag's default
		} else {
			// Fallback for standard flag package (shouldn't happen in Phase 1)
			fs.StringVar(&sf.Long, longName, defaultValue, usage)
			// Register short form handler to track if it was set
			fs.Func(shortName, usage+" (short form)", func(value string) error {
				sf.Short = value
				sf.shortSet = true
				return nil
			})
		}
	} else if longName != "" {
		fs.StringVar(&sf.Long, longName, defaultValue, usage)
	} else {
		// Only short form
		fs.Func(shortName, usage, func(value string) error {
			sf.Short = value
			sf.shortSet = true
			sf.Long = value // Also set Long for compatibility
			return nil
		})
	}
}

// RegisterBool registers both long and short form bool flags.
// At least one of longName or shortName must be provided.
func RegisterBool(fs *flag.FlagSet, bf *BoolFlag, longName, shortName string, usage string) {
	if longName == "" && shortName == "" {
		panic(fmt.Sprintf("RegisterBool: at least one of longName or shortName must be provided (got longName=%q, shortName=%q)", longName, shortName))
	}

	// TODO(Phase 2): Remove this reflection-based workaround
	// This is intentional technical debt from Phase 1 of the pflag migration.
	// Phase 1 (ticket: 250926-165945-phase1-pflag-basic-import-migration) replaced imports only.
	// Phase 2 (ticket: 250926-170130-phase2-pflag-proper-registration) will:
	//   - Remove all reflection-based code
	//   - Use pflag's BoolVarP directly
	//   - Properly handle flag precedence with pflag's native behavior
	// See docs/pflag-migration-phases.md for full migration plan
	//
	// Phase 1 pflag compatibility: Use pflag's BoolVarP when both are provided
	// This is a temporary fix for Phase 1 - will be properly refactored in Phase 2
	if longName != "" && shortName != "" {
		// BoolVarP is a pflag-specific method that registers both long and short forms
		// We use reflection to call it if available (when using pflag)
		if method := reflect.ValueOf(fs).MethodByName("BoolVarP"); method.IsValid() {
			// Call BoolVarP(p *bool, name, shorthand string, value bool, usage string)
			method.Call([]reflect.Value{
				reflect.ValueOf(&bf.Long),
				reflect.ValueOf(longName),
				reflect.ValueOf(shortName),
				reflect.ValueOf(false),
				reflect.ValueOf(usage),
			})
			// Note: With pflag, we can't track short vs long precedence in Phase 1
			// pflag's behavior is "last flag wins" which differs from our custom logic
			// This will be properly addressed in Phase 2
		} else {
			// Fallback for standard flag package (shouldn't happen in Phase 1)
			fs.BoolVar(&bf.Long, longName, false, usage)
			fs.BoolVar(&bf.Short, shortName, false, usage+" (short form)")
		}
	} else if longName != "" {
		fs.BoolVar(&bf.Long, longName, false, usage)
	} else {
		fs.BoolVar(&bf.Short, shortName, false, usage)
		// Also set Long for compatibility
		fs.BoolVar(&bf.Long, shortName, false, usage)
	}
}

// Value returns the resolved string value
// Phase 1 note: With pflag, "last flag wins" instead of "short takes precedence"
// This will be addressed properly in Phase 2
func (sf *StringFlag) Value() string {
	// With pflag's StringVarP, both forms write to sf.Long
	// Original behavior (short takes precedence) only works with standard flag package
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
