package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/lorudden/hemsida/internal/app"

	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/google/uuid"
)

const serviceName string = "hemsida"

func main() {

	ctx, flags := parseExternalConfig(context.Background(), app.DefaultFlags())

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

func parseExternalConfig(ctx context.Context, flags app.Flags) (context.Context, app.Flags) {

	flags[app.ControlPort] = env.GetVariableOrDefault(ctx, "CONTROL_PORT", flags[app.ControlPort])
	flags[app.ServicePort] = env.GetVariableOrDefault(ctx, "SERVICE_PORT", flags[app.ServicePort])
	flags[app.MediaDataPath] = env.GetVariableOrDefault(ctx, "MEDIA_DATA_PATH", flags[app.MediaDataPath])
	flags[app.ContentDataPath] = env.GetVariableOrDefault(ctx, "CONTENT_DATA_PATH", flags[app.ContentDataPath])

	apply := func(f app.Flag) func(string) error {
		return func(value string) error {
			flags[f] = value
			return nil
		}
	}

	// Allow command line arguments to override defaults and environment variables
	flag.BoolFunc("devmode", "enable devmode with fake backend data", apply(app.DevModeEnabled))
	flag.Func("listen", "network and address to listen to", apply(app.ListenAddress))
	flag.Func("controlport", "port number to bind to for the control interface", apply(app.ControlPort))
	flag.Func("port", "port number to bind to for the public interface", apply(app.ServicePort))
	flag.Func("web-assets", "path to web assets folder", apply(app.WebAssetPath))
	flag.Func("media-data", "path to the JSON media catalog directory", apply(app.MediaDataPath))
	flag.Func("content-data", "path to the JSON content catalog directory", apply(app.ContentDataPath))
	flag.Func("log-format", "choose to get log output in text or json format", apply(app.LogFormat))
	flag.Parse()

	return ctx, flags
}

func exitIf(err error, logger *slog.Logger, msg string, args ...any) {
	if err != nil {
		logger.With(args...).Error(msg, "err", err.Error())
		os.Exit(1)
	}
}
