package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	frontendtoolkit "github.com/diwise/frontend-toolkit"
	"github.com/diwise/frontend-toolkit/pkg/assets"
	"github.com/diwise/frontend-toolkit/pkg/locale"
	"github.com/diwise/frontend-toolkit/pkg/middleware/csp"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/google/uuid"

	"github.com/lorudden/hemsida/internal/pkg/application"
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

func Logger(ctx context.Context) func(http.Handler) http.Handler {

	log := logging.GetFromContext(ctx)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logging.NewContextWithLogger(r.Context(), log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RegisterHandlers(appContext context.Context, handler *http.ServeMux, assetLoader frontendtoolkit.AssetLoader, l10n frontendtoolkit.LocaleBundle, app application.App) error {

	version := uuid.NewString()

	mux := http.NewServeMux()

	assets.RegisterEndpoints(appContext, assetLoader, assets.WithMux(mux),
		assets.WithImmutableExpiry(48*time.Hour),
		assets.WithRedirect("/favicon.ico", "/icons/favicon.ico", http.StatusFound),
	)

	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		acceptLanguage := r.Header.Get("Accept-Language")
		localizer := l10n.For(acceptLanguage)

		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		home := components.StartPage(version, localizer, assetLoader.Load)
		home.Render(ctx, w)
	}))

	mux.Handle("GET /api", jsonapi.NewJSONAPIHandler(appContext))

	handler.Handle("GET /api/sse/{version}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetFromContext(ctx)

		flusher, ok := w.(http.Flusher)
		if !ok {
			logger.Warn("streaming not supported for this response writer")
			http.Error(w, "unable to start event stream", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		const eventFmt string = "event: %s\ndata: %s\n\n"

		logger.Info("comparing versions", "client", r.PathValue("version"), "mine", version)
		waitingForUpgrade := false

		if strings.Compare(r.PathValue("version"), version) != 0 {
			waitingForUpgrade = true

			logger.Info("sending upgrade to client")
			fmt.Fprintf(w, eventFmt, "upgrade2", version)
			flusher.Flush()

			time.Sleep(10 * time.Second)

			logger.Info("sending goodbye to client")
			fmt.Fprintf(w, eventFmt, "goodbye", version)
			flusher.Flush()
		} else {
			logger.Info("sse client successfully connected")
		}

		defer func() { logger.Info("exiting sse handler") }()

		tmr := time.NewTicker(time.Second)

		for {
			select {
			case t := <-tmr.C:
				if !waitingForUpgrade {
					fmt.Fprintf(w, eventFmt, "tick", t.Format(time.RFC3339Nano))
				}
			case <-ctx.Done():
				logger.Info("sse client connection closed")
				return
			case <-appContext.Done():
				logger.Info("we are shutting down")
				fmt.Fprintf(w, eventFmt, "goodbye", version)
				flusher.Flush()
				return
			}

			flusher.Flush()
		}

	}))

	handler.Handle("GET /", Logger(appContext)(csp.NewContentSecurityPolicy(csp.StrictDynamic())(mux)))

	return nil
}
