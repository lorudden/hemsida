package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lorudden/hemsida/internal/app"

	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/google/uuid"
)

const serviceName string = "hemsida"

func main() {
	ctx, flags := app.ResolveFlags(context.Background())

	serviceVersion := buildinfo.SourceVersion()
	if serviceVersion == "" || flags[app.DevModeEnabled] == "true" {
		serviceVersion = "develop" + "-" + uuid.NewString()
	}

	ctx, logger, cleanup := o11y.Init(ctx, serviceName, serviceVersion, flags[app.LogFormat])
	defer cleanup()

	cfg, err := app.NewConfig(ctx, flags)
	exitIf(err, logger, "failed to create application config")

	runner, err := app.Initialize(ctx, flags, cfg)
	exitIf(err, logger, "failed to initialize service runner")

	err = runner.Run(ctx)
	exitIf(err, logger, "failed to run service")
}

func exitIf(err error, logger *slog.Logger, msg string, args ...any) {
	if err != nil {
		logger.With(args...).Error(msg, "err", err.Error())
		os.Exit(1)
	}
}
