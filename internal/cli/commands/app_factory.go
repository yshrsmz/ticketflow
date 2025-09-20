package commands

import (
	"context"

	"github.com/yshrsmz/ticketflow/internal/cli"
)

// getAppWithFormat returns an App instance with the specified output format
func getAppWithFormat(ctx context.Context, format cli.OutputFormat) (*cli.App, error) {
	return cli.NewAppWithFormat(ctx, format)
}
