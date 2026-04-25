package catalog

import (
	"context"
	"testing"
)

func TestFileRepositoryListCollections(t *testing.T) {
	repository, err := NewFileRepository("testdata/repository")
	if err != nil {
		t.Fatalf("NewFileRepository() error = %v", err)
	}

	collections, err := repository.ListCollections(context.Background())
	if err != nil {
		t.Fatalf("ListCollections() error = %v", err)
	}
	if got := len(collections); got != 2 {
		t.Fatalf("len(collections) = %d, want 2", got)
	}
	if got := collections[0].ID; got != "news" {
		t.Fatalf("collections[0].ID = %q, want news", got)
	}
	if got := collections[1].ID; got != "pages" {
		t.Fatalf("collections[1].ID = %q, want pages", got)
	}
	if got := collections[1].Items[0].ID; got != "hamnforeningen" {
		t.Fatalf("collections[1].Items[0].ID = %q, want hamnforeningen", got)
	}
	if got := collections[1].Items[1].ID; got != "lorudden" {
		t.Fatalf("collections[1].Items[1].ID = %q, want lorudden", got)
	}
}
