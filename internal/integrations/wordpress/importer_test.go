package wordpress

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lorudden/hemsida/internal/content/catalog"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestImporterImportCollectionFileDryRun(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/pages.json"
	input := `{
  "id": "pages",
  "title": "Pages",
  "items": [
    {
      "id": "lorudden",
      "kind": "page",
      "title": "Lörudden",
      "source_page_url": "https://example.com/lorudden",
      "migration_status": "inventory"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	importer := NewImporter(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `<html><head><title>Lörudden</title><meta property="article:published_time" content="2024-06-20T10:30:00Z"></head><body><main class="entry-content"><p>Importerad ingress med mer innehall for sammanfattningen.</p><div class="hero"><img src="/hero.jpg" alt="Hero"></div><figure><img src="/detail.jpg" alt="Detail"><figcaption>Detalj</figcaption></figure></main></body></html>`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	result, err := importer.ImportCollectionFile(context.Background(), path, true)
	if err != nil {
		t.Fatalf("ImportCollectionFile() error = %v", err)
	}

	if got := result.EntriesUpdated; got != 1 {
		t.Fatalf("result.EntriesUpdated = %d, want 1", got)
	}

	collection, err := readCollectionFile(path)
	if err != nil {
		t.Fatalf("readCollectionFile() error = %v", err)
	}
	if got := collection.Items[0].MigrationStatus; got != "inventory" {
		t.Fatalf("dry run should not write file, got migration_status %q", got)
	}
}

func TestBuildImageCollections(t *testing.T) {
	entry := &catalog.Entry{ID: "lorudden", Slug: "lorudden", Title: "Lörudden"}
	collections := []catalog.ImageCollection{
		{Title: "Hero Images", Kind: "hero", Items: []catalog.ImageReference{{ID: "1", SourceURL: "https://example.com/hero.jpg"}}},
		{Title: "Gallery Images", Kind: "gallery", Items: []catalog.ImageReference{{ID: "2", SourceURL: "https://example.com/detail.jpg"}}},
	}

	built := buildImageCollections(entry, collections)
	if built[0].ID != "lorudden-hero" {
		t.Fatalf("built[0].ID = %q, want lorudden-hero", built[0].ID)
	}
	if built[1].ID != "lorudden-gallery-2" {
		t.Fatalf("built[1].ID = %q, want lorudden-gallery-2", built[1].ID)
	}
	if built[0].ImportMode != catalog.ImageImportModeExternal {
		t.Fatalf("built[0].ImportMode = %q", built[0].ImportMode)
	}
}

func TestImporterSetsPublishedAtAndCollections(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/pages.json"
	mediaRoot := filepath.Join(dir, "media")
	input := `{
  "id": "pages",
  "title": "Pages",
  "items": [
    {
      "id": "entry-1",
      "kind": "page",
      "title": "",
      "summary": "",
      "source_page_url": "https://example.com/post",
      "migration_status": "inventory"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	importer := NewImporterWithOptions(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://example.com/post":
			body := `<html><head><title>Ny post</title><meta property="article:published_time" content="2024-06-20T10:30:00Z"></head><body><main class="entry-content"><h1>Ny post</h1><p>Detta ar en importerad sammanfattning med tillrackligt mycket text.</p><div class="gallery"><img src="/1.jpg" alt="1"><img src="/2.jpg" alt="2"></div></main></body></html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		case "https://example.com/1.jpg", "https://example.com/2.jpg":
			header := make(http.Header)
			header.Set("Content-Type", "image/jpeg")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("jpeg-bits")),
				Header:     header,
			}, nil
		default:
			t.Fatalf("unexpected request URL %q", req.URL.String())
			return nil, nil
		}
	}), ImporterOptions{MediaRoot: mediaRoot})

	_, err := importer.ImportCollectionFile(context.Background(), path, false)
	if err != nil {
		t.Fatalf("ImportCollectionFile() error = %v", err)
	}

	collection, err := readCollectionFile(path)
	if err != nil {
		t.Fatalf("readCollectionFile() error = %v", err)
	}
	entry := collection.Items[0]
	if entry.PublishedAt == nil {
		t.Fatalf("entry.PublishedAt = nil")
	}
	if got := entry.PublishedAt.Format(time.RFC3339); got != "2024-06-20T10:30:00Z" {
		t.Fatalf("entry.PublishedAt = %q", got)
	}
	if got := len(entry.ImageCollections); got != 1 {
		t.Fatalf("len(entry.ImageCollections) = %d, want 1", got)
	}
	if got := entry.ImageImportMode; got != catalog.ImageImportModeDownloaded {
		t.Fatalf("entry.ImageImportMode = %q", got)
	}
	firstImage := entry.ImageCollections[0].Items[0]
	if got := firstImage.StoragePath; got == "" {
		t.Fatalf("first image StoragePath is empty")
	}
	if got := firstImage.MIMEType; got != "image/jpeg" {
		t.Fatalf("first image MIMEType = %q", got)
	}
	if _, err := os.Stat(filepath.Join(mediaRoot, filepath.FromSlash(firstImage.StoragePath))); err != nil {
		t.Fatalf("downloaded image missing: %v", err)
	}
}

func TestImporterExpandsNewsArchiveIntoEntries(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/news.json"
	mediaRoot := filepath.Join(dir, "media")
	input := `{
  "id": "news",
  "title": "News",
  "source_page_url": "https://example.com/news",
  "items": [
    {
      "id": "news-landing",
      "kind": "news",
      "title": "Nyheter",
      "planned_path": "/nyheter",
      "source_page_url": "https://example.com/news"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	importer := NewImporterWithOptions(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://example.com/news":
			body := `<html><body><div class="post type-post"><h2 class="entry-title"><a href="https://example.com/?p=2113">Brandövning</a></h2><div class="entry-meta"><span class="entry-date">juli 6, 2025</span></div><div class="entry-content"><p>Detta ar en nyhetssammanfattning med tillrackligt mycket text for att valjas.</p><figure class="wp-block-image"><a href="http://example.com/wp-content/uploads/brand.jpg"><img src="http://example.com/wp-content/uploads/brand-724x1024.jpg"></a></figure></div></div></body></html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		case "https://example.com/wp-content/uploads/brand.jpg":
			header := make(http.Header)
			header.Set("Content-Type", "image/jpeg; charset=binary")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("brand-image")),
				Header:     header,
			}, nil
		default:
			t.Fatalf("unexpected request URL %q", req.URL.String())
			return nil, nil
		}
	}), ImporterOptions{MediaRoot: mediaRoot})

	result, err := importer.ImportCollectionFile(context.Background(), path, false)
	if err != nil {
		t.Fatalf("ImportCollectionFile() error = %v", err)
	}
	if got := result.EntriesUpdated; got != 1 {
		t.Fatalf("result.EntriesUpdated = %d, want 1", got)
	}

	collection, err := readCollectionFile(path)
	if err != nil {
		t.Fatalf("readCollectionFile() error = %v", err)
	}
	if got := len(collection.Items); got != 1 {
		t.Fatalf("len(collection.Items) = %d, want 1", got)
	}
	if got := collection.Items[0].ID; got != "post-2113" {
		t.Fatalf("collection.Items[0].ID = %q", got)
	}
	if got := collection.Items[0].PlannedPath; got != "/nyheter/post-2113" {
		t.Fatalf("collection.Items[0].PlannedPath = %q", got)
	}
	if got := collection.Items[0].ImageImportMode; got != catalog.ImageImportModeDownloaded {
		t.Fatalf("collection.Items[0].ImageImportMode = %q", got)
	}
	if got := collection.Items[0].ImageCollections[0].Items[0].StoragePath; got == "" {
		t.Fatalf("collection.Items[0].ImageCollections[0].Items[0].StoragePath is empty")
	}
}
