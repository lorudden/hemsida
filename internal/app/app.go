package app

import (
	"context"

	"github.com/lorudden/hemsida/internal/content/media"
)

// Application represents the assembled application dependency graph.
type Application interface {
	MediaRepository() media.Repository
}

type application struct {
	mediaRepository media.Repository
}

func newApplication(_ context.Context, cfg *Config) (Application, error) {
	repository, err := media.NewFileRepository(cfg.MediaDataPath)
	if err != nil {
		return nil, err
	}

	return &application{
		mediaRepository: repository,
	}, nil
}

func (a *application) MediaRepository() media.Repository {
	return a.mediaRepository
}
