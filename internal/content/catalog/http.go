package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NewJSONAPIHandler exposes the page and news catalog as a JSON:API document.
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
			http.Error(w, "unable to load content catalog", http.StatusInternalServerError)
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
				itemRefs = append(itemRefs, resourceIdentifier{Type: "content-items", ID: item.ID})
				doc.Included = append(doc.Included, resourceObject{
					Type: "content-items",
					ID:   item.ID,
					Attributes: map[string]any{
						"title":            item.Title,
						"summary":          item.Summary,
						"kind":             item.Kind,
						"slug":             item.Slug,
						"source_page_url":  item.SourcePageURL,
						"planned_path":     item.PlannedPath,
						"migration_status": item.MigrationStatus,
						"visibility":       item.Visibility,
						"tags":             item.Tags,
						"published_at":     item.PublishedAt,
						"last_modified_at": item.LastModifiedAt,
					},
				})
			}

			doc.Data = append(doc.Data, resourceObject{
				Type: "content-collections",
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
			http.Error(w, "unable to encode content catalog", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(responseBody)))
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBody)
	})
}
