package jsonapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func NewJSONAPIHandler(appContext context.Context) http.HandlerFunc {

	type jsonMetaApi struct {
		Version string `json:"version"`
	}

	type jsonapiMeta struct {
		Api jsonMetaApi `json:"jsonapi"`
	}

	type jsonapiDataItem struct {
		Type       string            `json:"type"`
		ID         string            `json:"id"`
		Attributes map[string]string `json:"attributes"`
	}

	type jsonapiObj struct {
		Meta  jsonapiMeta `json:"meta"`
		Data  any         `json:"data,omitempty"`
		Error any         `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		obj := &jsonapiObj{
			Meta: jsonapiMeta{
				Api: jsonMetaApi{Version: "1.1"},
			},
		}

		data := make([]jsonapiDataItem, 0, 10)

		data = append(data, jsonapiDataItem{
			Type: "articles",
			ID:   "1",
			Attributes: map[string]string{
				"author": "kallek",
				"title":  "Välkommen till Midsommarfirande!",
			},
		})

		obj.Data = data

		responseBody, _ := json.Marshal(obj)

		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(responseBody)))
		w.Header().Add("Content-Type", "application/vnd.api+json")

		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}
}
