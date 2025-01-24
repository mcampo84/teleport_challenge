package main

import (
	"context"

	"go.uber.org/fx"

	"github.com/mcampo84/teleport_challenge/lib/cli"
	"github.com/mcampo84/teleport_challenge/lib/client"
)

func main() {
	app := fx.New(
		fx.Provide(
			client.GetDefaultConfig,
		),
		cli.Module,
		fx.Invoke(StartClient),
	)

	app.Run()
}

// StartClient gracefully starts and stops a client.
//
// Parameters:
//
//	c: The CLI interface
//  client: The gRPC client
//
// Returns:
//
//	An error if the client could not be started.
func StartClient(lc fx.Lifecycle, c *cli.CLI, client *client.Client) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Use a context without a timeout for the CLI
			go func() {
				if err := c.Start(context.Background()); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			client.GracefulStop()
			return nil
		},
	})
}
