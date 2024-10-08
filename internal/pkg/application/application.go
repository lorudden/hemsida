package application

import "context"

type App interface{}

type impl struct{}

func New(ctx context.Context) (App, error) {
	return &impl{}, nil
}
