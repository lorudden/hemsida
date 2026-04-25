package catalog

import (
	"net/http"

	platformapi "github.com/lorudden/hemsida/internal/platform/http/api"
)

// NewJSONAPIHandler exposes the page and news catalog as a JSON:API document.
func NewJSONAPIHandler(repository Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collections, err := repository.ListCollections(r.Context())
		if err != nil {
			http.Error(w, "unable to load content catalog", http.StatusInternalServerError)
			return
		}

		doc := platformapi.NewDocument()

		for _, collection := range collections {
			itemRefs := make([]platformapi.ResourceIdentifier, 0, len(collection.Items))
			for _, item := range collection.Items {
				itemRefs = append(itemRefs, platformapi.ResourceIdentifier{Type: "content-items", ID: item.ID})
				doc.Included = append(doc.Included, platformapi.ResourceObject{
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

			doc.Data = append(doc.Data, platformapi.ResourceObject{
				Type: "content-collections",
				ID:   collection.ID,
				Attributes: map[string]any{
					"title":           collection.Title,
					"description":     collection.Description,
					"source_page_url": collection.SourcePageURL,
				},
				Relationships: map[string]platformapi.RelationshipData{
					"items": {Data: itemRefs},
				},
			})
		}

		if err := platformapi.WriteDocument(w, doc); err != nil {
			http.Error(w, "unable to encode content catalog", http.StatusInternalServerError)
		}
	})
}
