package media

import (
	"net/http"

	platformapi "github.com/lorudden/hemsida/internal/platform/http/api"
)

// NewJSONAPIHandler exposes the media catalog as a JSON:API document.
func NewJSONAPIHandler(repository Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collections, err := repository.ListCollections(r.Context())
		if err != nil {
			http.Error(w, "unable to load media catalog", http.StatusInternalServerError)
			return
		}

		doc := platformapi.NewDocument()

		for _, collection := range collections {
			itemRefs := make([]platformapi.ResourceIdentifier, 0, len(collection.Items))
			for _, item := range collection.Items {
				itemRefs = append(itemRefs, platformapi.ResourceIdentifier{Type: "media-items", ID: item.ID})
				doc.Included = append(doc.Included, platformapi.ResourceObject{
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

			doc.Data = append(doc.Data, platformapi.ResourceObject{
				Type: "media-collections",
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
			http.Error(w, "unable to encode media catalog", http.StatusInternalServerError)
		}
	})
}
