package cli

import (
	"go.uber.org/fx"

	"github.com/mcampo84/teleport_challenge/lib/client"
)

var Module = fx.Module(
	"CLI",
	fx.Provide(
		NewCLI,
	),
	client.Module,
)
