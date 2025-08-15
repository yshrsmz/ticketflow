package commands

import (
	"context"

	"github.com/yshrsmz/ticketflow/internal/cli"
)

// getApp returns an App instance, using test factory if set
func getApp(ctx context.Context) (*cli.App, error) {
	if testAppFactory != nil {
		return testAppFactory(ctx)
	}
	return cli.NewApp(ctx)
}
