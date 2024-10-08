package api

import (
	"context"
	"net/http"

	"github.com/lorudden/hemsida/internal/pkg/application"
	"github.com/lorudden/hemsida/internal/pkg/presentation/web/components"
)

type Api interface {
	Router() *http.ServeMux
}

type impl struct {
	mux *http.ServeMux
}

func New(ctx context.Context, mux *http.ServeMux, app application.App) (Api, error) {

	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		home := components.Home()
		home.Render(ctx, w)
	}))

	return &impl{
		mux: mux,
	}, nil
}

func (a *impl) Router() *http.ServeMux {
	return a.mux
}
