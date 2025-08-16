package cli

import (
	"context"
	"os"
)

// NewAppWithFormat creates a new App with the specified output format
// This ensures the App is properly initialized with the correct writers from the start
func NewAppWithFormat(ctx context.Context, format OutputFormat) (*App, error) {
	return NewAppWithOptions(ctx,
		WithOutput(NewOutputWriter(os.Stdout, os.Stderr, format)),
		WithStatusWriter(NewStatusWriter(os.Stdout, format)),
	)
}

// WithOutput sets the output writer
func WithOutput(output *OutputWriter) AppOption {
	return func(app *App) {
		app.Output = output
	}
}

// WithStatusWriter sets the status writer
func WithStatusWriter(writer StatusWriter) AppOption {
	return func(app *App) {
		app.StatusWriter = writer
	}
}