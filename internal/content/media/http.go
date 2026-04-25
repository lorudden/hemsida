package media

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NewJSONAPIHandler exposes the media catalog as a JSON:API document.
func NewJSONAPIHandler(repository Repository) http.Handler {
	type resourceIdentifier struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}

	type relationshipData struct {
		Data []resourceIdentifier `json:"data"`
	}

	type resourceObject struct {
		Type          string                      `json:"type"`
		ID            string                      `json:"id"`
		Attributes    map[string]any              `json:"attributes,omitempty"`
		Relationships map[string]relationshipData `json:"relationships,omitempty"`
	}

	type document struct {
		JSONAPI struct {
			Version string `json:"version"`
		} `json:"jsonapi"`
		Data     []resourceObject `json:"data"`
		Included []resourceObject `json:"included,omitempty"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collections, err := repository.ListCollections(r.Context())
		if err != nil {
			http.Error(w, "unable to load media catalog", http.StatusInternalServerError)
			return
		}

		doc := document{
			Data:     make([]resourceObject, 0, len(collections)),
			Included: make([]resourceObject, 0),
		}
		doc.JSONAPI.Version = "1.1"

		for _, collection := range collections {
			itemRefs := make([]resourceIdentifier, 0, len(collection.Items))
			for _, item := range collection.Items {
				itemRefs = append(itemRefs, resourceIdentifier{Type: "media-items", ID: item.ID})
				doc.Included = append(doc.Included, resourceObject{
					Type: "media-items",
					ID:   item.ID,
					Attributes: map[string]any{
						"title":            item.Title,
						"description":      item.Description,
						"kind":             item.Kind,
						"source_url":       item.SourceURL,
						"source_page_url":  item.SourcePageURL,
						"storage_path":     item.StoragePath,
						"mime_type":        item.MIMEType,
						"migration_status": item.MigrationStatus,
						"visibility":       item.Visibility,
						"attribution":      item.Attribution,
						"tags":             item.Tags,
						"published_at":     item.PublishedAt,
						"last_modified_at": item.LastModifiedAt,
					},
				})
			}

			doc.Data = append(doc.Data, resourceObject{
				Type: "media-collections",
				ID:   collection.ID,
				Attributes: map[string]any{
					"title":           collection.Title,
					"description":     collection.Description,
					"source_page_url": collection.SourcePageURL,
				},
				Relationships: map[string]relationshipData{
					"items": {Data: itemRefs},
				},
			})
		}

		responseBody, err := json.Marshal(doc)
		if err != nil {
			http.Error(w, "unable to encode media catalog", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(responseBody)))
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBody)
	})
}
