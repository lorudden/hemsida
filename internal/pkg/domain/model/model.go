package model

type Identity interface {
	ID() string
}

type Document interface {
	Identity
}

type WebPage interface {
	Identity
}

type ImageGallery interface {
	Identity
}

type Image interface {
	Identity
}

type News interface {
	Identity
}

type WebLink interface {
	Identity
}
