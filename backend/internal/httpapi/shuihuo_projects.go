package httpapi

import (
	"net/http"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// registerShuihuoProjectCompatRoutes keeps the legacy shared-work library
// contract available while the browser still reads its project cards through
// /api/platform-projects. The source of truth is the Go batch store.
func registerShuihuoProjectCompatRoutes(root *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store) {
	if store == nil {
		return
	}
	root.Handle("/api/shuihuo-production/projects", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		projects := make([]map[string]any, 0, len(batches))
		for _, batch := range batches {
			createdAt := batch.CreatedAt.UTC().Format(time.RFC3339)
			updatedAt := batch.UpdatedAt.UTC().Format(time.RFC3339)
			status := "pending"
			if len(batch.Books) > 0 {
				status = "confirmed"
			}
			projects = append(projects, map[string]any{
				"id":                 batch.ID,
				"name":               batch.Title,
				"createdAt":          createdAt,
				"updatedAt":          updatedAt,
				"segmentationStatus": status,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
	})))
}
