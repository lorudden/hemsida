package api

import (
	"context"
	"net/http"

	"github.com/lorudden/hemsida/internal/pkg/application"
	"github.com/lorudden/hemsida/internal/pkg/presentation/web/components"
)

func RegisterHandlers(ctx context.Context, mux *http.ServeMux, app application.App) error {

	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		home := components.Home()
		home.Render(ctx, w)
	}))

	return nil
}
