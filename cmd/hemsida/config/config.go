package config

import (
	"context"
	"net/http"

	"github.com/diwise/service-chassis/pkg/infrastructure/servicerunner"
	"github.com/lorudden/hemsida/internal/pkg/application"
	"github.com/lorudden/hemsida/internal/pkg/presentation/api"
)

type FlagType int
type FlagMap map[FlagType]string

const (
	ListenAddress FlagType = iota
	ServicePort
	ControlPort

	WebAssetPath

	DevModeEnabled

	/*webAssetPath

	appRoot

	oauth2RealmURL
	oauth2ClientID
	oauth2ClientSecret*/
)

func DefaultFlags() FlagMap {
	return FlagMap{
		ListenAddress: "",
		ServicePort:   "8080",
		ControlPort:   "",

		DevModeEnabled: "false",
	}
}

func New(ctx context.Context, flags FlagMap) (*AppConfig, error) {
	return &AppConfig{}, nil
}

func Initialize(ctx context.Context, flags FlagMap, cfg *AppConfig) (servicerunner.Runner[AppConfig], error) {
	var err error
	cfg.app, err = application.New(ctx)
	if err != nil {
		return nil, err
	}

	_, runner := servicerunner.New(ctx, *cfg,
		ifnot(flags[ControlPort] == "",
			webserver("control", listen(flags[ListenAddress]), port(flags[ControlPort]), pprof()),
		),
		webserver("public", listen(flags[ListenAddress]), port(flags[ServicePort]),
			muxinit(func(ctx context.Context, identifier string, port string, svcCfg *AppConfig, handler *http.ServeMux) error {
				if err = api.RegisterHandlers(ctx, handler, svcCfg.app); err != nil {
					return err
				}

				return nil
			}),
		))

	return runner, nil
}

type AppConfig struct {
	app application.App
}

var webserver = servicerunner.WithHTTPServeMux[AppConfig]
var muxinit = servicerunner.OnMuxInit[AppConfig]
var listen = servicerunner.WithListenAddr[AppConfig]
var port = servicerunner.WithPort[AppConfig]
var ifnot = servicerunner.IfNot[AppConfig]
var pprof = servicerunner.WithPPROF[AppConfig]
