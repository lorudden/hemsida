package media

import "context"

// Repository loads media collections from the configured backing store.
type Repository interface {
	ListCollections(context.Context) ([]Collection, error)
}
