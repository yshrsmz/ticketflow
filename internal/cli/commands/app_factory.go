package commands

import (
	"context"

	"github.com/yshrsmz/ticketflow/internal/cli"
)

// getAppWithFormat returns an App instance with the specified output format
func getAppWithFormat(ctx context.Context, format cli.OutputFormat) (*cli.App, error) {
	if testAppFactory != nil {
		// For tests, still use the test factory
		// Tests will need to be updated to handle format properly
		return testAppFactory(ctx)
	}
	return cli.NewAppWithFormat(ctx, format)
}
