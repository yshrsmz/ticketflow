package main

import (
	"context"
	"flag"
	"strings"
	"testing"
)

func TestParseAndExecute(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "simple command without flags",
			cmd: Command{
				Name: "test",
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					return nil
				},
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "command with required arguments",
			cmd: Command{
				Name:         "test",
				MinArgs:      1,
				MinArgsError: "missing test argument",
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					return nil
				},
			},
			args:    []string{},
			wantErr: true,
			errMsg:  "missing test argument",
		},
		{
			name: "command with required arguments provided",
			cmd: Command{
				Name:    "test",
				MinArgs: 1,
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					if fs.Arg(0) != "arg1" {
						t.Errorf("expected first arg to be 'arg1', got %s", fs.Arg(0))
					}
					return nil
				},
			},
			args:    []string{"arg1"},
			wantErr: false,
		},
		{
			name: "command with flags",
			cmd: Command{
				Name: "test",
				SetupFlags: func(fs *flag.FlagSet) interface{} {
					name := fs.String("name", "default", "Name flag")
					age := fs.Int("age", 0, "Age flag")
					return &struct {
						name *string
						age  *int
					}{name, age}
				},
				Execute: func(ctx context.Context, fs *flag.FlagSet, cmdFlags interface{}) error {
					flags := cmdFlags.(*struct {
						name *string
						age  *int
					})
					if *flags.name != "test-name" {
						t.Errorf("expected name to be 'test-name', got %s", *flags.name)
					}
					if *flags.age != 25 {
						t.Errorf("expected age to be 25, got %d", *flags.age)
					}
					return nil
				},
			},
			args:    []string{"-name", "test-name", "-age", "25"},
			wantErr: false,
		},
		{
			name: "command with custom validation",
			cmd: Command{
				Name: "test",
				Validate: func(fs *flag.FlagSet, flags interface{}) error {
					// Custom validation that always fails
					return flag.ErrHelp
				},
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					return nil
				},
			},
			args:    []string{},
			wantErr: true,
			errMsg:  "flag: help requested",
		},
		{
			name: "command with generic MinArgs error",
			cmd: Command{
				Name:    "test",
				MinArgs: 2,
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					return nil
				},
			},
			args:    []string{"arg1"},
			wantErr: true,
			errMsg:  "missing required arguments for test command",
		},
		{
			name: "command with context cancellation",
			cmd: Command{
				Name: "test",
				Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						return nil
					}
				},
			},
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := parseAndExecute(ctx, tt.cmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseAndExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseAndExecute() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseAndExecuteWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := Command{
		Name: "test",
		Execute: func(ctx context.Context, fs *flag.FlagSet, flags interface{}) error {
			// This should still be called even with cancelled context
			// as we don't check context before execution
			return nil
		},
	}

	// The command should execute successfully even with cancelled context
	// because parseAndExecute doesn't check context status
	err := parseAndExecute(ctx, cmd, []string{})
	if err != nil {
		t.Errorf("parseAndExecute() with cancelled context should not fail, got error: %v", err)
	}
}