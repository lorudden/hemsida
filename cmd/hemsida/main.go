package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/lorudden/hemsida/cmd/hemsida/config"

	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/google/uuid"
)

const serviceName string = "hemsida"

func main() {

	ctx, flags := parseExternalConfig(context.Background(), config.DefaultFlags())

	serviceVersion := buildinfo.SourceVersion()
	if serviceVersion == "" || flags[config.DevModeEnabled] == "true" {
		serviceVersion = "develop" + "-" + uuid.NewString()
	}

	ctx, logger, cleanup := o11y.Init(ctx, serviceName, serviceVersion)
	defer cleanup()

	cfg, err := config.New(ctx, flags)
	exitIf(err, logger, "failed to create application config")

	runner, err := config.Initialize(ctx, flags, cfg)
	exitIf(err, logger, "failed to initialize service runner")

	err = runner.Run(ctx)
	exitIf(err, logger, "failed to run service")
}

func parseExternalConfig(ctx context.Context, flags config.FlagMap) (context.Context, config.FlagMap) {

	apply := func(f config.FlagType) func(string) error {
		return func(value string) error {
			flags[f] = value
			return nil
		}
	}

	// Allow command line arguments to override defaults and environment variables
	flag.BoolFunc("devmode", "enable devmode with fake backend data", apply(config.DevModeEnabled))
	flag.Func("web-assets", "path to web assets folder", apply(config.WebAssetPath))
	flag.Parse()

	return ctx, flags
}

func exitIf(err error, logger *slog.Logger, msg string, args ...any) {
	if err != nil {
		logger.With(args...).Error(msg, "err", err.Error())
		os.Exit(1)
	}
}
