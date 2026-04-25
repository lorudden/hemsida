package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ResourceIdentifier identifies a related JSON:API resource.
type ResourceIdentifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RelationshipData holds relationship identifiers for a resource.
type RelationshipData struct {
	Data []ResourceIdentifier `json:"data"`
}

// ResourceObject is a generic JSON:API resource object.
type ResourceObject struct {
	Type          string                      `json:"type"`
	ID            string                      `json:"id"`
	Attributes    map[string]any              `json:"attributes,omitempty"`
	Relationships map[string]RelationshipData `json:"relationships,omitempty"`
}

// Document is a lightweight JSON:API document for collection endpoints.
type Document struct {
	JSONAPI struct {
		Version string `json:"version"`
	} `json:"jsonapi"`
	Data     []ResourceObject `json:"data"`
	Included []ResourceObject `json:"included,omitempty"`
}

// NewDocument creates a JSON:API 1.1 document.
func NewDocument() Document {
	doc := Document{
		Data:     make([]ResourceObject, 0),
		Included: make([]ResourceObject, 0),
	}
	doc.JSONAPI.Version = "1.1"
	return doc
}

// WriteDocument marshals and writes a JSON:API document to the response.
func WriteDocument(w http.ResponseWriter, doc any) error {
	responseBody, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(responseBody)))
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseBody)
	return err
}
