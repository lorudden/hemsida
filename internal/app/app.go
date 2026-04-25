package app

import (
	"context"

	"github.com/lorudden/hemsida/internal/content/catalog"
	"github.com/lorudden/hemsida/internal/content/media"
)

// Application represents the assembled application dependency graph.
type Application interface {
	ContentRepository() catalog.Repository
	MediaRepository() media.Repository
}

type application struct {
	contentRepository catalog.Repository
	mediaRepository   media.Repository
}

func newApplication(_ context.Context, cfg *Config) (Application, error) {
	contentRepository, err := catalog.NewFileRepository(cfg.ContentDataPath)
	if err != nil {
		return nil, err
	}

	mediaRepository, err := media.NewFileRepository(cfg.MediaDataPath)
	if err != nil {
		return nil, err
	}

	return &application{
		contentRepository: contentRepository,
		mediaRepository:   mediaRepository,
	}, nil
}

func (a *application) ContentRepository() catalog.Repository {
	return a.contentRepository
}

func (a *application) MediaRepository() media.Repository {
	return a.mediaRepository
}
