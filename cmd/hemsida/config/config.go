package config

import (
	"context"
	"net/http"

	"github.com/diwise/service-chassis/pkg/infrastructure/servicerunner"
	"github.com/lorudden/hemsida/internal/pkg/application"
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

		DevModeEnabled: "false",
	}
}

func New(ctx context.Context, flags Flags) (*AppData, error) {
	return &AppData{}, nil
}

func Initialize(ctx context.Context, flags Flags, cfg *AppData) (servicerunner.Runner[AppData], error) {
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
			muxinit(func(ctx context.Context, identifier string, port string, svcCfg *AppData, handler *http.ServeMux) error {
				if err = api.RegisterHandlers(ctx, handler, svcCfg.app); err != nil {
					return err
				}

				return nil
			}),
		))

	return runner, nil
}

type AppData struct {
	app application.App
}

var webserver = servicerunner.WithHTTPServeMux[AppData]
var muxinit = servicerunner.OnMuxInit[AppData]
var listen = servicerunner.WithListenAddr[AppData]
var port = servicerunner.WithPort[AppData]
var ifnot = servicerunner.IfNot[AppData]
var pprof = servicerunner.WithPPROF[AppData]
