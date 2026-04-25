package media

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/lorudden/hemsida/internal/platform/filejson"
)

// FileRepository reads media collections from JSON files in a directory.
type FileRepository struct {
	root string
}

// NewFileRepository creates a repository that reads one collection per JSON file.
func NewFileRepository(root string) (*FileRepository, error) {
	if err := filejson.RequirePath(root, "media"); err != nil {
		return nil, err
	}

	return &FileRepository{root: root}, nil
}

// ListCollections returns all media collections sorted by title and item title.
func (r *FileRepository) ListCollections(context.Context) ([]Collection, error) {
	collections, err := filejson.LoadDirectory(r.root, "media", validateCollection)
	if err != nil {
		return nil, err
	}

	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Title < collections[j].Title
	})

	for i := range collections {
		sort.Slice(collections[i].Items, func(left, right int) bool {
			return collections[i].Items[left].Title < collections[i].Items[right].Title
		})
	}

	return collections, nil
}

func validateCollection(path string, collection Collection) error {
	if strings.TrimSpace(collection.ID) == "" {
		return fmt.Errorf("media collection %q is missing id", path)
	}

	if strings.TrimSpace(collection.Title) == "" {
		return fmt.Errorf("media collection %q is missing title", path)
	}

	for _, item := range collection.Items {
		if strings.TrimSpace(item.ID) == "" {
			return fmt.Errorf("media collection %q contains item with missing id", path)
		}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("media collection %q contains item %q with missing title", path, item.ID)
		}
		if strings.TrimSpace(string(item.Kind)) == "" {
			return fmt.Errorf("media collection %q contains item %q with missing kind", path, item.ID)
		}
	}

	return nil
}
