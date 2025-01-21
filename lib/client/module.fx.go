package client

import "go.uber.org/fx"

var Module = fx.Module(
	"Client",
	fx.Provide(
		NewClient,
	),
)
