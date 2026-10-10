package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
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
		switch r.Method {
		case http.MethodGet:
			if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
				writeShuihuoProject(w, r, store, owner, id)
				return
			}
			batches, err := store.ListBatches(r.Context(), owner)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			projects := make([]map[string]any, 0, len(batches))
			for _, batch := range batches {
				projects = append(projects, shuihuoProject(batch))
			}
			writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
		case http.MethodPost:
			var input struct {
				Name           string `json:"name"`
				ProductionMode string `json:"productionMode,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			name := strings.TrimSpace(input.Name)
			if name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
				return
			}
			batch, err := store.CreateBatch(r.Context(), owner, batchfactoryv11.CreateBatchInput{
				Title: name,
				InitialBatchSettings: batchfactoryv11.SettingsPatch{
					"shuihuoProject": json.RawMessage("true"),
				},
				// Keep one editable source book so the legacy source PUT can update
				// a newly-created project before any segmentation work starts.
				Books: []batchfactoryv11.CreateBookInput{{Title: name}},
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			project := shuihuoProject(batch)
			if input.ProductionMode != "" {
				project["productionMode"] = input.ProductionMode
			}
			writeJSON(w, http.StatusCreated, project)
		case http.MethodDelete:
			id := strings.TrimSpace(r.URL.Query().Get("id"))
			if id == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
				return
			}
			if err := store.DeleteBatch(r.Context(), owner, id); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})))

	root.Handle("/api/shuihuo-production/projects/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/projects/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project id is required"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, parts[0])
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, shuihuoProject(batch))
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			if err := store.DeleteBatch(r.Context(), owner, batch.ID); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		if len(parts) == 2 && parts[1] == "source" && (r.Method == http.MethodPut || r.Method == http.MethodPost) {
			var input struct {
				SourceText       string `json:"sourceText"`
				SegmentationMode string `json:"segmentationMode,omitempty"`
				SourceTitle      string `json:"sourceTitle,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.TrimSpace(input.SourceText) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			book := batch.Books[0]
			updated, err := store.CaptureBookSource(r.Context(), owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{
				SourceText: strings.TrimSpace(input.SourceText), ReplaceSource: true, SourceTitle: input.SourceTitle, ExpectedRevision: book.Revision,
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			batch.Books[0] = updated
			project := shuihuoProject(batch)
			project["segmentationMode"] = input.SegmentationMode
			writeJSON(w, http.StatusOK, project)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project route not found"})
	})))
}

func isShuihuoProjectBatch(batch batchfactoryv11.Batch) bool {
	value, ok := batch.SettingsState.Patch["shuihuoProject"]
	return ok && strings.TrimSpace(string(value)) == "true"
}

func shuihuoProject(batch batchfactoryv11.Batch) map[string]any {
	createdAt := batch.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := batch.UpdatedAt.UTC().Format(time.RFC3339)
	status := "pending"
	var sourceText string
	if len(batch.Books) > 0 {
		sourceText = batch.Books[0].SourceText
		if strings.TrimSpace(sourceText) != "" {
			status = "confirmed"
		}
	}
	return map[string]any{
		"id": batch.ID, "name": batch.Title, "createdAt": createdAt, "updatedAt": updatedAt,
		"sourceText": sourceText, "segmentationStatus": status, "segments": []any{}, "assets": []any{},
		"books": batch.Books,
	}
}

func writeShuihuoProject(w http.ResponseWriter, r *http.Request, store batchfactoryv11.Store, owner, id string) {
	batch, err := store.GetBatch(r.Context(), owner, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, shuihuoProject(batch))
}
