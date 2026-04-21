package api

import (
	"context"
	"net/http"
	"time"

	frontendtoolkit "github.com/diwise/frontend-toolkit"
	"github.com/diwise/frontend-toolkit/pkg/assets"
	"github.com/diwise/frontend-toolkit/pkg/locale"
	"github.com/diwise/frontend-toolkit/pkg/middleware/csp"
	"github.com/diwise/service-chassis/pkg/infrastructure/net/http/requestlogger"
	"github.com/diwise/service-chassis/pkg/infrastructure/net/http/router"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/google/uuid"

	"github.com/lorudden/hemsida/internal/app"
	"github.com/lorudden/hemsida/internal/pkg/presentation/api/handlers"
	"github.com/lorudden/hemsida/internal/pkg/presentation/api/jsonapi"
	"github.com/lorudden/hemsida/internal/pkg/presentation/web/components"
)

func NewAssetLoader(ctx context.Context, assetPath string) (frontendtoolkit.AssetLoader, error) {
	return assets.NewLoader(ctx,
		assets.BasePath(assetPath), assets.Exclude("/l10n"),
		assets.Logger(logging.GetFromContext(ctx)),
	)
}

func NewLocaleBundle(ctx context.Context, assetPath string, languages []string) (frontendtoolkit.LocaleBundle, error) {
	l10n := locale.NewLocalizer(assetPath, languages...)
	return l10n, nil
}

func RegisterHandlers(appContext context.Context, handler *http.ServeMux, assetLoader frontendtoolkit.AssetLoader, l10n frontendtoolkit.LocaleBundle, app app.App) error {

	version := uuid.NewString()

	logger := logging.GetFromContext(appContext)

	r := router.New(handler)

	assets.RegisterEndpoints(appContext, assetLoader, assets.WithMux(handler),
		assets.WithImmutableExpiry(48*time.Hour),
		assets.WithRedirect("/favicon.ico", "/icons/favicon.ico", http.StatusFound),
	)

	r.Use(requestlogger.New(logger))
	r.Use(csp.NewContentSecurityPolicy(csp.StrictDynamic()))

	r.Get("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		acceptLanguage := r.Header.Get("Accept-Language")
		localizer := l10n.For(acceptLanguage)

		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		home := components.StartPage(version, localizer, assetLoader.Load)
		home.Render(ctx, w)
	}))

	r.Get("/api", jsonapi.NewJSONAPIHandler(appContext))
	r.Get("/api/sse/{version}", handlers.NewSSEHandler(appContext, version))

	return nil
}
