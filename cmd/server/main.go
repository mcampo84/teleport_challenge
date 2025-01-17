package main

import (
	"context"
	"log"

	"go.uber.org/fx"

	jobmanager "github.com/mcampo84/teleport_challenge/lib/job_manager"
	"github.com/mcampo84/teleport_challenge/lib/server"
)

func main() {
	app := fx.New(
		server.Module,
		jobmanager.Module,
		fx.Invoke(StartGRPCServer),
	)

	app.Run()
}

func StartGRPCServer(lc fx.Lifecycle, server *server.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Listen(); err != nil {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.GracefulStop()
			return nil
		},
	})
}
