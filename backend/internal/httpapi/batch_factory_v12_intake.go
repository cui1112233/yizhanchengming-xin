package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

// batchShuihuoProjectLister lets the V12 batch list read the "shuihuo project"
// flag for every batch in one query. Stores that do not implement it keep the
// previous per-batch fallback.
type batchShuihuoProjectLister interface {
	ListShuihuoProjectBatchIDs(context.Context, string) (map[string]bool, error)
}

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
		var batches []batchfactoryv11.Batch
		var err error
		if lister, ok := store.(batchSummaryIndexLister); ok {
			batches, err = lister.ListBatchSummaryIndex(r.Context(), owner)
		} else {
			batches, err = store.ListBatches(r.Context(), owner)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		summaries := make([]map[string]any, 0, len(batches))
		for _, batch := range batches {
			// The public 3001 library shows existing batches here, including
			// batches that were first created from Shuihuo.  Hiding them made the
			// 4001 batch-factory page look empty even though V11 could read them.
			// Keep the summary as a read-only index; detailed ownership checks stay
			// in GetBatch and the V11 routes.
			books := make([]map[string]any, 0, len(batch.Books))
			for _, book := range batch.Books {
				books = append(books, map[string]any{"id": book.ID, "bookId": book.BookID, "title": book.Title, "platform": book.Platform})
			}
			item := map[string]any{
				"id": batch.ID, "title": batch.Title, "createdAt": batch.CreatedAt,
				"updatedAt": batch.UpdatedAt, "revision": batch.Revision,
				"bookCount": len(books), "books": books,
			}
			if strings.TrimSpace(batch.ProjectCoverArtifactID) != "" {
				item["coverMedia"] = map[string]any{"kind": "image", "url": "/api/batch-factory/v12/batches/" + url.PathEscape(batch.ID) + "/merge-cover/" + url.PathEscape(batch.ProjectCoverArtifactID)}
			} else if strings.TrimSpace(batch.ProjectCoverMediaURL) != "" {
				item["coverMedia"] = map[string]any{"kind": "video", "url": batch.ProjectCoverMediaURL}
			} else if strings.TrimSpace(batch.ProjectCoverImageURL) != "" {
				item["coverMedia"] = map[string]any{"kind": "image", "url": batch.ProjectCoverImageURL}
			}
			item["productionMode"] = batchProductionMode(batch)
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
		writeJSON(w, http.StatusOK, map[string]any{"batches": filterShuihuoProjectBatches(r.Context(), store, owner, batches)})
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
			"redirectTo": "/batch-factory?intake=" + intake.ID,
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

// filterShuihuoProjectBatches hides batches that the batch settings mark as
// Shuihuo projects. When the store supports a bulk flag query this is one
// statement; otherwise (and on query failure) it keeps the original per-batch
// behaviour, including keeping batches whose detail cannot be read.
func filterShuihuoProjectBatches(ctx context.Context, store batchfactoryv11.Store, owner string, batches []batchfactoryv11.Batch) []batchfactoryv11.Batch {
	filtered := make([]batchfactoryv11.Batch, 0, len(batches))
	if lister, ok := store.(batchShuihuoProjectLister); ok {
		flags, err := lister.ListShuihuoProjectBatchIDs(ctx, owner)
		if err == nil {
			for _, batch := range batches {
				if !flags[batch.ID] {
					filtered = append(filtered, batch)
				}
			}
			return filtered
		}
	}
	for _, batch := range batches {
		detail, detailErr := store.GetBatch(ctx, owner, batch.ID)
		if detailErr != nil || !isShuihuoProjectBatch(detail) {
			filtered = append(filtered, batch)
		}
	}
	return filtered
}

// batchProductionMode 让项目卡片显示“漫画/视频”。摘要索引会带上推断结果；
// 走内存库或详情列表时退回读取批次与书籍设置里的 productionMode。
func batchProductionMode(batch batchfactoryv11.Batch) string {
	if mode := strings.ToLower(strings.TrimSpace(batch.ProductionMode)); mode == "comic" || mode == "video" {
		return mode
	}
	patches := []batchfactoryv11.SettingsPatch{batch.SettingsState.Patch}
	for _, book := range batch.Books {
		patches = append(patches, book.SettingsState.Patch)
		if strings.EqualFold(strings.TrimSpace(book.Mode), "comic") {
			return "comic"
		}
	}
	for _, patch := range patches {
		raw, ok := patch["productionMode"]
		if !ok {
			continue
		}
		var mode string
		if json.Unmarshal(raw, &mode) == nil && strings.EqualFold(strings.TrimSpace(mode), "comic") {
			return "comic"
		}
	}
	return "video"
}
