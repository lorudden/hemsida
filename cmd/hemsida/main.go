package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/google/uuid"
	"github.com/lorudden/hemsida/internal/pkg/application"
	"github.com/lorudden/hemsida/internal/pkg/presentation/api"
)

const serviceName string = "hemsida"

func main() {
	serviceVersion := buildinfo.SourceVersion()
	if serviceVersion == "" {
		serviceVersion = "develop" + "-" + uuid.NewString()
	}

	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion)
	defer cleanup()

	mux := http.NewServeMux()

	webapi, _, err := initialize(ctx, mux, "")
	if err != nil {
		fatal(ctx, "failed to initialize service", err)
	}

	apiPort := "8080"
	webServer := &http.Server{Addr: ":" + apiPort, Handler: webapi.Router()}

	logger.Info("starting to listen for incoming connections", "port", apiPort)

	go func() {
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fatal(ctx, "failed to start request router", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigChan

	logger.Info("received signal", "signal", s)

	err = webServer.Shutdown(ctx)
	if err != nil {
		logger.Error("failed to shutdown web server", "err", err.Error())
	}

	logger.Info("shutting down")
}

func fatal(ctx context.Context, msg string, err error) {
	logging.GetFromContext(ctx).Error(msg, "err", err.Error())
	os.Exit(1)
}

func initialize(ctx context.Context, mux *http.ServeMux, _ string) (api_ api.Api, app application.App, err error) {
	app, err = application.New(ctx)
	if err != nil {
		return
	}

	api_, err = api.New(ctx, mux, app)
	if err != nil {
		return
	}

	return api_, app, nil
}
