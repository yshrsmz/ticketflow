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
// The short form uses an empty default to detect if it was explicitly set.
func RegisterString(fs *flag.FlagSet, sf *StringFlag, longName, shortName, defaultValue, usage string) {
	sf.Long = defaultValue
	
	// Register long form with default
	fs.StringVar(&sf.Long, longName, defaultValue, usage)
	
	// Register short form with custom handler to track if it was set
	fs.Func(shortName, usage+" (short form)", func(value string) error {
		sf.Short = value
		sf.shortSet = true
		return nil
	})
}

// RegisterBool registers both long and short form bool flags
func RegisterBool(fs *flag.FlagSet, bf *BoolFlag, longName, shortName string, usage string) {
	fs.BoolVar(&bf.Long, longName, false, usage)
	fs.BoolVar(&bf.Short, shortName, false, usage+" (short form)")
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

// MergeFlags is a helper that can be used in Validate to resolve all flag pairs at once
type MergeFlags struct {
	Format *StringFlag
	Parent *StringFlag  
	Force  *BoolFlag
}

// ResolveAll resolves all flag pairs and returns their final values
func (m *MergeFlags) ResolveAll() (format string, parent string, force bool) {
	if m.Format != nil {
		format = m.Format.Value()
	}
	if m.Parent != nil {
		parent = m.Parent.Value()
	}
	if m.Force != nil {
		force = m.Force.Value()
	}
	return
}