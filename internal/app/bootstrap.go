package app

import (
	"context"
	"net/http"

	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/servicerunner"
	platformapi "github.com/lorudden/hemsida/internal/platform/http/api"
)

type runtimeState struct {
	cancelContext context.CancelFunc
}

// Initialize wires the application into the service-chassis runner.
func Initialize(ctx context.Context, flags Flags, cfg *Config) (servicerunner.Runner[Config], error) {
	var err error

	state := runtimeState{}
	_, err = newApplication(ctx)
	if err != nil {
		return nil, err
	}

	assetLoader, err := platformapi.NewAssetLoader(ctx, flags[WebAssetPath])
	if err != nil {
		return nil, err
	}

	l10n, err := platformapi.NewLocaleBundle(ctx, flags[WebAssetPath], []string{"en", "sv"})
	if err != nil {
		return nil, err
	}

	ctx, state.cancelContext = context.WithCancel(ctx)

	_, runner := servicerunner.New(ctx, *cfg,
		ifnot(flags[ControlPort] == "",
			webserver("control", listen(flags[ListenAddress]), port(flags[ControlPort]), pprof()),
		),
		webserver("public", listen(flags[ListenAddress]), port(flags[ServicePort]),
			muxinit(func(ctx context.Context, identifier string, port string, svcCfg *Config, handler *http.ServeMux) error {
				if err = platformapi.RegisterHandlers(ctx, handler, assetLoader, l10n); err != nil {
					return err
				}

				logging.GetFromContext(ctx).Info("server running", "port", port)

				return nil
			}),
		),
		onshutdown(func(ctx context.Context, svcCfg *Config) error {
			logging.GetFromContext(ctx).Info("shutting down ...")
			state.cancelContext()
			return nil
		}),
	)

	return runner, nil
}

var webserver = servicerunner.WithHTTPServeMux[Config]
var muxinit = servicerunner.OnMuxInit[Config]
var listen = servicerunner.WithListenAddr[Config]
var port = servicerunner.WithPort[Config]
var ifnot = servicerunner.IfNot[Config]
var pprof = servicerunner.WithPPROF[Config]
var onshutdown = servicerunner.OnShutdown[Config]
