package model

// Identity is implemented by entities with stable identifiers.
type Identity interface {
	ID() string
}

// Document represents a published document.
type Document interface {
	Identity
}

// WebPage represents a renderable web page.
type WebPage interface {
	Identity
}

// ImageGallery represents a gallery of related images.
type ImageGallery interface {
	Identity
}

// Image represents an image asset.
type Image interface {
	Identity
}

// News represents a published news item.
type News interface {
	Identity
}

// WebLink represents a curated outbound link.
type WebLink interface {
	Identity
}
