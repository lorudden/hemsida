package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileRepository reads content collections from JSON files in a directory.
type FileRepository struct {
	root string
}

// NewFileRepository creates a repository that reads one collection per JSON file.
func NewFileRepository(root string) (*FileRepository, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("content repository path is required")
	}

	return &FileRepository{root: root}, nil
}

// ListCollections returns all content collections sorted by title and item title.
func (r *FileRepository) ListCollections(context.Context) ([]Collection, error) {
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return nil, fmt.Errorf("read content repository: %w", err)
	}

	collections := make([]Collection, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		collection, err := r.readCollection(filepath.Join(r.root, entry.Name()))
		if err != nil {
			return nil, err
		}

		collections = append(collections, collection)
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

func (r *FileRepository) readCollection(path string) (Collection, error) {
	file, err := os.Open(path)
	if err != nil {
		return Collection{}, fmt.Errorf("open content collection %q: %w", path, err)
	}
	defer file.Close()

	payload, err := io.ReadAll(file)
	if err != nil {
		return Collection{}, fmt.Errorf("read content collection %q: %w", path, err)
	}

	var collection Collection
	if err := json.Unmarshal(payload, &collection); err != nil {
		return Collection{}, fmt.Errorf("decode content collection %q: %w", path, err)
	}

	if strings.TrimSpace(collection.ID) == "" {
		return Collection{}, fmt.Errorf("content collection %q is missing id", path)
	}

	if strings.TrimSpace(collection.Title) == "" {
		return Collection{}, fmt.Errorf("content collection %q is missing title", path)
	}

	for _, item := range collection.Items {
		if strings.TrimSpace(item.ID) == "" {
			return Collection{}, fmt.Errorf("content collection %q contains item with missing id", path)
		}
		if strings.TrimSpace(item.Title) == "" {
			return Collection{}, fmt.Errorf("content collection %q contains item %q with missing title", path, item.ID)
		}
		if strings.TrimSpace(string(item.Kind)) == "" {
			return Collection{}, fmt.Errorf("content collection %q contains item %q with missing kind", path, item.ID)
		}
	}

	return collection, nil
}
