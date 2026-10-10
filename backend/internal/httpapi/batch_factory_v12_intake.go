package httpapi

import (
	"net/http"
	"net/url"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

// registerV12IntakeRoutes keeps the public novel-fetch handoff on the same
// durable Go batch store as the V11 workbench routes. The browser uses V12 for
// the handoff, while the production workbench continues reading the intake
// through its V11-compatible contract.
func registerV12IntakeRoutes(mux *http.ServeMux, store batchfactoryv11.Store) {
	mux.HandleFunc("GET /api/batch-factory/v12/batches/summary", func(w http.ResponseWriter, r *http.Request) {
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
		summaries := make([]map[string]any, 0, len(batches))
		for _, batch := range batches {
			detail, detailErr := store.GetBatch(r.Context(), owner, batch.ID)
			if detailErr == nil && isShuihuoProjectBatch(detail) {
				continue
			}
			books := make([]map[string]any, 0, len(batch.Books))
			for _, book := range batch.Books {
				books = append(books, map[string]any{"id": book.ID, "bookId": book.BookID, "title": book.Title, "platform": book.Platform})
			}
			item := map[string]any{
				"id": batch.ID, "title": batch.Title, "createdAt": batch.CreatedAt,
				"updatedAt": batch.UpdatedAt, "revision": batch.Revision,
				"bookCount": len(books), "books": books,
			}
			if strings.TrimSpace(batch.ProjectCoverJobID) != "" {
				item["coverMedia"] = map[string]any{"kind": "image", "url": "/api/batch-factory/v11/batches/" + url.PathEscape(batch.ID) + "/merge-cover/" + url.PathEscape(batch.ProjectCoverJobID)}
			}
			summaries = append(summaries, item)
		}
		writeJSON(w, http.StatusOK, map[string]any{"batches": summaries})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/batches", func(w http.ResponseWriter, r *http.Request) {
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
		filtered := make([]batchfactoryv11.Batch, 0, len(batches))
		for _, batch := range batches {
			detail, detailErr := store.GetBatch(r.Context(), owner, batch.ID)
			if detailErr != nil || !isShuihuoProjectBatch(detail) {
				filtered = append(filtered, batch)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"batches": filtered})
	})
	mux.HandleFunc("POST /api/batch-factory/v12/intakes/novel-fetch", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.NovelFetchIntakeInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if len(input.Books) == 0 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		intake, err := store.CreateIntake(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"intake":     intake,
			"redirectTo": "/shuihuo-production?intake=" + intake.ID,
		})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/intakes/{intakeId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		intake, err := store.GetIntake(r.Context(), owner, r.PathValue("intakeId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"intake": intake})
	})
}
