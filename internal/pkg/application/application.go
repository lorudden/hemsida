package application

import "context"

type App any

type impl struct{}

func New(context.Context) (App, error) {
	return &impl{}, nil
}
