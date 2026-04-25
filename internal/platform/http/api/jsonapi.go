package api

import (
	"context"
	"net/http"
)

// NewJSONAPIHandler returns the JSON:API endpoint handler.
func NewJSONAPIHandler(context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doc := NewDocument()
		doc.Data = append(doc.Data, ResourceObject{
			Type: "articles",
			ID:   "1",
			Attributes: map[string]any{
				"author": "kallek",
				"title":  "Välkommen till Midsommarfirande!",
			},
		})

		if err := WriteDocument(w, doc); err != nil {
			http.Error(w, "unable to encode jsonapi response", http.StatusInternalServerError)
		}
	}
}
