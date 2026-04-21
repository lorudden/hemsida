package config

import (
	"context"
	"net/http"

	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/servicerunner"
	"github.com/lorudden/hemsida/internal/app"
	"github.com/lorudden/hemsida/internal/pkg/presentation/api"
)

type Flag int
type Flags map[Flag]string

const (
	ListenAddress Flag = iota
	ServicePort
	ControlPort

	WebAssetPath

	DevModeEnabled

	LogFormat

	/*webAssetPath

	appRoot

	oauth2RealmURL
	oauth2ClientID
	oauth2ClientSecret*/
)

func DefaultFlags() Flags {
	return Flags{
		ListenAddress: "",
		ServicePort:   "8080",
		ControlPort:   "",

		WebAssetPath: "/opt/lorudden/assets",

		DevModeEnabled: "false",
		LogFormat:      "json",
	}
}

func New(ctx context.Context, flags Flags) (*AppData, error) {
	return &AppData{}, nil
}

func Initialize(ctx context.Context, flags Flags, cfg *AppData) (servicerunner.Runner[AppData], error) {
	var err error
	cfg.app, err = app.New(ctx)
	if err != nil {
		return nil, err
	}

	assetLoader, err := api.NewAssetLoader(ctx, flags[WebAssetPath])
	if err != nil {
		return nil, err
	}

	l10n, err := api.NewLocaleBundle(ctx, flags[WebAssetPath], []string{"en", "sv"})
	if err != nil {
		return nil, err
	}

	ctx, cfg.cancelContext = context.WithCancel(ctx)

	_, runner := servicerunner.New(ctx, *cfg,
		ifnot(flags[ControlPort] == "",
			webserver("control", listen(flags[ListenAddress]), port(flags[ControlPort]), pprof()),
		),
		webserver("public", listen(flags[ListenAddress]), port(flags[ServicePort]),
			muxinit(func(ctx context.Context, identifier string, port string, svcCfg *AppData, handler *http.ServeMux) error {
				if err = api.RegisterHandlers(ctx, handler, assetLoader, l10n, svcCfg.app); err != nil {
					return err
				}

				logging.GetFromContext(ctx).Info("server running", "port", port)

				return nil
			}),
		),
		onshutdown(func(ctx context.Context, svcCfg *AppData) error {
			logging.GetFromContext(ctx).Info("shutting down ...")
			svcCfg.cancelContext()
			return nil
		}),
	)

	return runner, nil
}

type AppData struct {
	app app.App

	cancelContext context.CancelFunc
}

var webserver = servicerunner.WithHTTPServeMux[AppData]
var muxinit = servicerunner.OnMuxInit[AppData]
var listen = servicerunner.WithListenAddr[AppData]
var port = servicerunner.WithPort[AppData]
var ifnot = servicerunner.IfNot[AppData]
var pprof = servicerunner.WithPPROF[AppData]
var onshutdown = servicerunner.OnShutdown[AppData]
