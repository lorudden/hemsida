package media

import "time"

// Kind identifies the type of published media in the migration catalog.
type Kind string

const (
	KindDocument Kind = "document"
	KindImage    Kind = "image"
)

// Item describes a single migratable media asset.
type Item struct {
	ID              string     `json:"id"`
	Kind            Kind       `json:"kind"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	SourceURL       string     `json:"source_url,omitempty"`
	SourcePageURL   string     `json:"source_page_url,omitempty"`
	StoragePath     string     `json:"storage_path,omitempty"`
	MIMEType        string     `json:"mime_type,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	LastModifiedAt  *time.Time `json:"last_modified_at,omitempty"`
	MigrationStatus string     `json:"migration_status,omitempty"`
	Visibility      string     `json:"visibility,omitempty"`
	Attribution     string     `json:"attribution,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
}

// Collection groups related media items into a file-backed JSON document.
type Collection struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	SourcePageURL string `json:"source_page_url,omitempty"`
	Items         []Item `json:"items"`
}
