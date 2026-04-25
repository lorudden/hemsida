package catalog

import "context"

// Repository loads page and news collections from the configured backing store.
type Repository interface {
	ListCollections(context.Context) ([]Collection, error)
}
