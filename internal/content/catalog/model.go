package catalog

import "time"

// Kind identifies the type of visible website content in the migration catalog.
type Kind string

const (
	KindPage Kind = "page"
	KindNews Kind = "news"
)

// Entry describes a single visible page or post we want to migrate.
type Entry struct {
	ID              string     `json:"id"`
	Kind            Kind       `json:"kind"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug,omitempty"`
	Summary         string     `json:"summary,omitempty"`
	SourcePageURL   string     `json:"source_page_url,omitempty"`
	PlannedPath     string     `json:"planned_path,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	LastModifiedAt  *time.Time `json:"last_modified_at,omitempty"`
	MigrationStatus string     `json:"migration_status,omitempty"`
	Visibility      string     `json:"visibility,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
}

// Collection groups related visible content into a file-backed JSON manifest.
type Collection struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	SourcePageURL string  `json:"source_page_url,omitempty"`
	Items         []Entry `json:"items"`
}
