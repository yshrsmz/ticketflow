package commands

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringFlag(t *testing.T) {
	tests := []struct {
		name          string
		setupFlags    func(*flag.FlagSet, *StringFlag)
		args          []string
		expectedValue string
	}{
		{
			name: "only long form provided",
			setupFlags: func(fs *flag.FlagSet, sf *StringFlag) {
				RegisterString(fs, sf, "format", "o", "text", "Output format")
			},
			args:          []string{"-format", "json"},
			expectedValue: "json",
		},
		{
			name: "only short form provided",
			setupFlags: func(fs *flag.FlagSet, sf *StringFlag) {
				RegisterString(fs, sf, "format", "o", "text", "Output format")
			},
			args:          []string{"-o", "json"},
			expectedValue: "json",
		},
		{
			name: "both provided - short takes precedence",
			setupFlags: func(fs *flag.FlagSet, sf *StringFlag) {
				RegisterString(fs, sf, "format", "o", "text", "Output format")
			},
			args:          []string{"-format", "json", "-o", "yaml"},
			expectedValue: "yaml",
		},
		{
			name: "neither provided - uses default",
			setupFlags: func(fs *flag.FlagSet, sf *StringFlag) {
				RegisterString(fs, sf, "format", "o", "text", "Output format")
			},
			args:          []string{},
			expectedValue: "text",
		},
		{
			name: "short form explicitly set to default value",
			setupFlags: func(fs *flag.FlagSet, sf *StringFlag) {
				RegisterString(fs, sf, "format", "o", "text", "Output format")
			},
			args:          []string{"-format", "json", "-o", "text"},
			expectedValue: "text", // Short form wins even when it's the default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			sf := &StringFlag{}

			tt.setupFlags(fs, sf)

			err := fs.Parse(tt.args)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedValue, sf.Value())
		})
	}
}

func TestBoolFlag(t *testing.T) {
	tests := []struct {
		name          string
		setupFlags    func(*flag.FlagSet, *BoolFlag)
		args          []string
		expectedValue bool
	}{
		{
			name: "only long form true",
			setupFlags: func(fs *flag.FlagSet, bf *BoolFlag) {
				RegisterBool(fs, bf, "force", "f", "Force operation")
			},
			args:          []string{"-force"},
			expectedValue: true,
		},
		{
			name: "only short form true",
			setupFlags: func(fs *flag.FlagSet, bf *BoolFlag) {
				RegisterBool(fs, bf, "force", "f", "Force operation")
			},
			args:          []string{"-f"},
			expectedValue: true,
		},
		{
			name: "both true - OR operation",
			setupFlags: func(fs *flag.FlagSet, bf *BoolFlag) {
				RegisterBool(fs, bf, "force", "f", "Force operation")
			},
			args:          []string{"-force", "-f"},
			expectedValue: true,
		},
		{
			name: "neither provided - false",
			setupFlags: func(fs *flag.FlagSet, bf *BoolFlag) {
				RegisterBool(fs, bf, "force", "f", "Force operation")
			},
			args:          []string{},
			expectedValue: false,
		},
		{
			name: "only long true, short false",
			setupFlags: func(fs *flag.FlagSet, bf *BoolFlag) {
				RegisterBool(fs, bf, "force", "f", "Force operation")
			},
			args:          []string{"-force=true", "-f=false"},
			expectedValue: true, // OR operation means true if either is true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			bf := &BoolFlag{}

			tt.setupFlags(fs, bf)

			err := fs.Parse(tt.args)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedValue, bf.Value())
		})
	}
}

// TestRegisterPanics tests that RegisterString and RegisterBool panic with helpful messages
func TestRegisterPanics(t *testing.T) {
	t.Run("RegisterString panics with empty names", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		sf := &StringFlag{}

		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Contains(t, r.(string), "longName=\"\"")
			assert.Contains(t, r.(string), "shortName=\"\"")
		}()

		RegisterString(fs, sf, "", "", "default", "usage")
	})

	t.Run("RegisterBool panics with empty names", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		bf := &BoolFlag{}

		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Contains(t, r.(string), "longName=\"\"")
			assert.Contains(t, r.(string), "shortName=\"\"")
		}()

		RegisterBool(fs, bf, "", "", "usage")
	})
}

// TestOldVsNewApproach demonstrates the improvement
func TestOldVsNewApproach(t *testing.T) {
	t.Run("old approach - bug with default values", func(t *testing.T) {
		// This simulates the old bug where --format json was ignored
		// when -o had default value "text"
		fs := flag.NewFlagSet("old", flag.ContinueOnError)

		var format, formatShort string
		fs.StringVar(&format, "format", "text", "Output format")
		fs.StringVar(&formatShort, "o", "text", "Output format (short)")

		// User provides --format json (but not -o)
		err := fs.Parse([]string{"-format", "json"})
		assert.NoError(t, err)

		// Old normalization logic (buggy)
		// This was the bug: formatShort is "text" (default) and would override
		if formatShort != "" && formatShort != "text" {
			format = formatShort
		}

		// Bug: format is still "json" because we skip default "text"
		// But this required a hack to check for default value
		assert.Equal(t, "json", format)
	})

	t.Run("new approach - clean solution", func(t *testing.T) {
		fs := flag.NewFlagSet("new", flag.ContinueOnError)

		sf := &StringFlag{}
		RegisterString(fs, sf, "format", "o", "text", "Output format")

		// User provides --format json (but not -o)
		err := fs.Parse([]string{"-format", "json"})
		assert.NoError(t, err)

		// Clean: Value() method handles it correctly
		assert.Equal(t, "json", sf.Value())
	})

	t.Run("new approach - handles explicit default correctly", func(t *testing.T) {
		fs := flag.NewFlagSet("new", flag.ContinueOnError)

		sf := &StringFlag{}
		RegisterString(fs, sf, "format", "o", "text", "Output format")

		// User explicitly provides both: --format json -o text
		err := fs.Parse([]string{"-format", "json", "-o", "text"})
		assert.NoError(t, err)

		// Correctly returns "text" because -o was explicitly set
		assert.Equal(t, "text", sf.Value())
	})
}
