package app

import "context"

// Application represents the assembled application dependency graph.
type Application any

type application struct{}

func newApplication(context.Context) (Application, error) {
	return &application{}, nil
}
