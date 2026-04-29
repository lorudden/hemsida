package catalog

import "time"

// Kind identifies the type of visible website content in the migration catalog.
type Kind string

const (
	KindPage Kind = "page"
	KindNews Kind = "news"
)

// ImageImportMode describes how image assets are handled during migration.
type ImageImportMode string

const (
	ImageImportModeExternal   ImageImportMode = "external-link"
	ImageImportModeDownloaded ImageImportMode = "downloaded"
)

// ImageReference points at an image that remains hosted on the source site for now.
type ImageReference struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	SourceURL   string `json:"source_url"`
	StoragePath string `json:"storage_path,omitempty"`
	MIMEType    string `json:"mime_type,omitempty"`
	AltText     string `json:"alt_text,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Attribution string `json:"attribution,omitempty"`
}

// ImageCollection groups related images so the new site can build carousels and galleries.
type ImageCollection struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Kind       string           `json:"kind,omitempty"`
	ImportMode ImageImportMode  `json:"import_mode,omitempty"`
	Items      []ImageReference `json:"items"`
}

// Entry describes a single visible page or post we want to migrate.
type Entry struct {
	ID               string            `json:"id"`
	Kind             Kind              `json:"kind"`
	Title            string            `json:"title"`
	Slug             string            `json:"slug,omitempty"`
	Summary          string            `json:"summary,omitempty"`
	SourcePageURL    string            `json:"source_page_url,omitempty"`
	PlannedPath      string            `json:"planned_path,omitempty"`
	PublishedAt      *time.Time        `json:"published_at,omitempty"`
	LastModifiedAt   *time.Time        `json:"last_modified_at,omitempty"`
	MigrationStatus  string            `json:"migration_status,omitempty"`
	Visibility       string            `json:"visibility,omitempty"`
	ImageImportMode  ImageImportMode   `json:"image_import_mode,omitempty"`
	ImageCollections []ImageCollection `json:"image_collections,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
}

// Collection groups related visible content into a file-backed JSON manifest.
type Collection struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	SourcePageURL string  `json:"source_page_url,omitempty"`
	Items         []Entry `json:"items"`
}
