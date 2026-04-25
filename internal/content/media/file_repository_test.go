package media

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
	if got := collections[0].ID; got != "documents" {
		t.Fatalf("collections[0].ID = %q, want documents", got)
	}
	if got := collections[1].ID; got != "gallery" {
		t.Fatalf("collections[1].ID = %q, want gallery", got)
	}
	if got := collections[0].Items[0].ID; got != "charter" {
		t.Fatalf("collections[0].Items[0].ID = %q, want charter", got)
	}
	if got := collections[0].Items[1].ID; got != "harbour-report" {
		t.Fatalf("collections[0].Items[1].ID = %q, want harbour-report", got)
	}
}
