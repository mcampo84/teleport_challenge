package jobmanager

import "go.uber.org/fx"

var Module = fx.Module(
	"Job Manager",
	fx.Provide(
		NewJobManager,
	),
)
