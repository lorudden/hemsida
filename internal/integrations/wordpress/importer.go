package wordpress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lorudden/hemsida/internal/content/catalog"
)

// HTTPClient captures the subset of http.Client used by the importer.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ImporterOptions describes optional importer behavior.
type ImporterOptions struct {
	MediaRoot string
}

// Importer enriches content manifests with data fetched from WordPress pages.
type Importer struct {
	client     HTTPClient
	mediaStore *legacyMediaStore
}

// NewImporter creates a WordPress manifest importer.
func NewImporter(client HTTPClient) *Importer {
	return NewImporterWithOptions(client, ImporterOptions{})
}

// NewImporterWithOptions creates a WordPress manifest importer with optional media download support.
func NewImporterWithOptions(client HTTPClient, options ImporterOptions) *Importer {
	return &Importer{
		client:     client,
		mediaStore: newLegacyMediaStore(client, options.MediaRoot),
	}
}

// ImportResult describes the work performed for a collection file.
type ImportResult struct {
	Path           string
	EntriesUpdated int
}

// ImportCollectionFile reads, enriches, and optionally writes a single collection manifest.
func (i *Importer) ImportCollectionFile(ctx context.Context, path string, dryRun bool) (ImportResult, error) {
	collection, err := readCollectionFile(path)
	if err != nil {
		return ImportResult{}, err
	}

	if isNewsCollection(collection) {
		return i.importNewsCollectionFile(ctx, path, collection, dryRun)
	}

	updated := 0
	for index := range collection.Items {
		changed, err := i.enrichEntry(ctx, &collection.Items[index], dryRun)
		if err != nil {
			return ImportResult{}, fmt.Errorf("import %q entry %q: %w", path, collection.Items[index].ID, err)
		}
		if changed {
			updated++
		}
	}

	if !dryRun {
		if err := writeCollectionFile(path, collection); err != nil {
			return ImportResult{}, err
		}
	}

	return ImportResult{
		Path:           path,
		EntriesUpdated: updated,
	}, nil
}

func (i *Importer) importNewsCollectionFile(ctx context.Context, path string, collection catalog.Collection, dryRun bool) (ImportResult, error) {
	sourceURL := strings.TrimSpace(collection.SourcePageURL)
	if sourceURL == "" && len(collection.Items) > 0 {
		sourceURL = strings.TrimSpace(collection.Items[0].SourcePageURL)
	}
	if sourceURL == "" {
		return ImportResult{}, nil
	}

	archiveEntries, err := i.fetchArchiveEntries(ctx, sourceURL)
	if err != nil {
		return ImportResult{}, fmt.Errorf("import archive %q: %w", path, err)
	}

	basePath := "/nyheter"
	if len(collection.Items) > 0 && strings.TrimSpace(collection.Items[0].PlannedPath) != "" {
		basePath = strings.TrimSpace(collection.Items[0].PlannedPath)
	}

	items := make([]catalog.Entry, 0, len(archiveEntries))
	updated := 0
	for _, archiveEntry := range archiveEntries {
		entry := buildNewsEntry(archiveEntry, basePath)
		changed, err := i.prepareEntryImages(ctx, &entry, dryRun)
		if err != nil {
			return ImportResult{}, fmt.Errorf("prepare news entry %q images: %w", entry.ID, err)
		}
		if changed {
			updated++
		}
		items = append(items, entry)
	}
	collection.Items = items

	if !dryRun {
		if err := writeCollectionFile(path, collection); err != nil {
			return ImportResult{}, err
		}
	}

	return ImportResult{
		Path:           path,
		EntriesUpdated: updated,
	}, nil
}

func (i *Importer) enrichEntry(ctx context.Context, entry *catalog.Entry, dryRun bool) (bool, error) {
	if strings.TrimSpace(entry.SourcePageURL) == "" {
		return i.prepareEntryImages(ctx, entry, dryRun)
	}

	page, err := i.fetchPage(ctx, entry.SourcePageURL)
	if err != nil {
		return false, err
	}

	changed := false

	if shouldReplaceTitle(entry.Title, page.Title) {
		entry.Title = page.Title
		changed = true
	}

	if shouldReplaceSummary(entry.Summary, page.Summary) {
		entry.Summary = page.Summary
		changed = true
	}

	if entry.PublishedAt == nil && page.PublishedAt != nil {
		entry.PublishedAt = page.PublishedAt
		changed = true
	}

	if len(page.ImageCollections) > 0 {
		collections := buildImageCollections(entry, page.ImageCollections)
		if !equalImageCollections(entry.ImageCollections, collections) {
			entry.ImageCollections = collections
			changed = true
		}
	}

	imageChanged, err := i.prepareEntryImages(ctx, entry, dryRun)
	if err != nil {
		return false, err
	}
	if imageChanged {
		changed = true
	}

	if entry.MigrationStatus == "inventory" {
		entry.MigrationStatus = "imported"
		changed = true
	}

	return changed, nil
}

func (i *Importer) prepareEntryImages(ctx context.Context, entry *catalog.Entry, dryRun bool) (bool, error) {
	if len(entry.ImageCollections) == 0 {
		return false, nil
	}

	collections, mode, changed, err := i.prepareImageCollections(ctx, entry.ImageCollections, dryRun)
	if err != nil {
		return false, err
	}

	if changed || !equalImageCollections(entry.ImageCollections, collections) {
		entry.ImageCollections = collections
		changed = true
	}

	if mode == "" {
		mode = catalog.ImageImportModeExternal
	}
	if entry.ImageImportMode != mode {
		entry.ImageImportMode = mode
		changed = true
	}

	return changed, nil
}

func (i *Importer) fetchPage(ctx context.Context, sourceURL string) (ParsedPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return ParsedPage{}, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return ParsedPage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ParsedPage{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return ParsePage(resp.Body, sourceURL)
}

func (i *Importer) fetchArchiveEntries(ctx context.Context, sourceURL string) ([]ParsedArchiveEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return ParseNewsArchive(resp.Body, sourceURL)
}

func readCollectionFile(path string) (catalog.Collection, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return catalog.Collection{}, fmt.Errorf("read collection file %q: %w", path, err)
	}

	var collection catalog.Collection
	if err := json.Unmarshal(payload, &collection); err != nil {
		return catalog.Collection{}, fmt.Errorf("decode collection file %q: %w", path, err)
	}

	return collection, nil
}

func writeCollectionFile(path string, collection catalog.Collection) error {
	payload, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("encode collection file %q: %w", path, err)
	}

	payload = append(payload, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare collection directory for %q: %w", path, err)
	}

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write collection file %q: %w", path, err)
	}

	return nil
}

func buildImageCollections(entry *catalog.Entry, collections []catalog.ImageCollection) []catalog.ImageCollection {
	built := make([]catalog.ImageCollection, 0, len(collections))
	baseID := entry.ID
	if strings.TrimSpace(entry.Slug) != "" {
		baseID = entry.Slug
	}

	for index, collection := range collections {
		normalized := collection
		normalized.ImportMode = catalog.ImageImportModeExternal
		if strings.TrimSpace(normalized.Kind) == "" {
			if len(normalized.Items) > 1 {
				normalized.Kind = "carousel"
			} else {
				normalized.Kind = "inline"
			}
		}
		if strings.TrimSpace(normalized.Title) == "" {
			normalized.Title = entry.Title + " bilder"
		}

		suffix := slugify(normalized.Kind)
		if suffix == "" {
			suffix = "images"
		}
		if index > 0 {
			normalized.ID = slugify(fmt.Sprintf("%s-%s-%d", baseID, suffix, index+1))
		} else {
			normalized.ID = slugify(fmt.Sprintf("%s-%s", baseID, suffix))
		}

		built = append(built, normalized)
	}

	return built
}

func (i *Importer) prepareImageCollections(ctx context.Context, collections []catalog.ImageCollection, dryRun bool) ([]catalog.ImageCollection, catalog.ImageImportMode, bool, error) {
	prepared := make([]catalog.ImageCollection, 0, len(collections))
	mode := catalog.ImageImportModeExternal
	changed := false

	for _, collection := range collections {
		normalized := collection
		items := make([]catalog.ImageReference, 0, len(collection.Items))

		for _, item := range collection.Items {
			preparedItem := item
			if !dryRun && i.mediaStore != nil {
				downloadedItem, err := i.mediaStore.Download(ctx, item)
				if err != nil {
					return nil, "", false, fmt.Errorf("download image %q: %w", item.SourceURL, err)
				}
				preparedItem = downloadedItem
			}

			if preparedItem.StoragePath != item.StoragePath || preparedItem.MIMEType != item.MIMEType {
				changed = true
			}
			items = append(items, preparedItem)
		}

		normalized.Items = items
		normalized.ImportMode = detectCollectionImportMode(items)
		if normalized.ImportMode == catalog.ImageImportModeDownloaded {
			mode = catalog.ImageImportModeDownloaded
		}
		if normalized.ImportMode != collection.ImportMode {
			changed = true
		}
		prepared = append(prepared, normalized)
	}

	return prepared, mode, changed, nil
}

func buildNewsEntry(parsed ParsedArchiveEntry, basePath string) catalog.Entry {
	entry := catalog.Entry{
		ID:              parsed.Slug,
		Kind:            catalog.KindNews,
		Title:           parsed.Title,
		Slug:            parsed.Slug,
		Summary:         parsed.Summary,
		SourcePageURL:   parsed.SourcePageURL,
		PlannedPath:     joinPath(basePath, parsed.Slug),
		PublishedAt:     parsed.PublishedAt,
		MigrationStatus: "imported",
		Visibility:      "public",
		ImageImportMode: catalog.ImageImportModeExternal,
		Tags:            []string{"migration", "news"},
	}
	if len(parsed.ImageCollections) > 0 {
		entry.ImageCollections = buildImageCollections(&entry, parsed.ImageCollections)
	}
	return entry
}

func detectCollectionImportMode(items []catalog.ImageReference) catalog.ImageImportMode {
	for _, item := range items {
		if strings.TrimSpace(item.StoragePath) != "" {
			return catalog.ImageImportModeDownloaded
		}
	}
	return catalog.ImageImportModeExternal
}

func shouldReplaceTitle(existing string, imported string) bool {
	imported = strings.TrimSpace(imported)
	if imported == "" {
		return false
	}
	return strings.TrimSpace(existing) == "" || looksInventoried(existing)
}

func shouldReplaceSummary(existing string, imported string) bool {
	imported = strings.TrimSpace(imported)
	if imported == "" {
		return false
	}
	return strings.TrimSpace(existing) == "" || looksInventoried(existing)
}

func looksInventoried(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return strings.Contains(value, "första") ||
		strings.Contains(value, "startmanifest") ||
		strings.Contains(value, "naturlig första") ||
		strings.Contains(value, "informationssida") ||
		strings.Contains(value, "föreningssida") ||
		strings.Contains(value, "bör mappas") ||
		strings.Contains(value, "landningssida")
}

func equalImageCollections(left []catalog.ImageCollection, right []catalog.ImageCollection) bool {
	leftPayload, _ := json.Marshal(left)
	rightPayload, _ := json.Marshal(right)
	return string(leftPayload) == string(rightPayload)
}

func isNewsCollection(collection catalog.Collection) bool {
	return strings.EqualFold(strings.TrimSpace(collection.ID), "news")
}

func joinPath(basePath string, slug string) string {
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")
	slug = strings.Trim(slug, "/")
	if basePath == "" {
		basePath = "/nyheter"
	}
	if slug == "" {
		return basePath
	}
	return basePath + "/" + slug
}

// DefaultHTTPClient returns an HTTP client suitable for import runs.
func DefaultHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// DefaultCollectionFiles returns the conventional content manifest files to import.
func DefaultCollectionFiles(contentDir string) []string {
	files := []string{
		filepath.Join(contentDir, "pages.json"),
		filepath.Join(contentDir, "news.json"),
	}

	sort.Strings(files)
	return files
}
