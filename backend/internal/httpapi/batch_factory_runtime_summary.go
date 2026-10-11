package httpapi

import (
	"context"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

func registerRuntimeSummaryRoute(mux *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, stageService *batchfactoryv11.BookStageService, production *batchfactoryv11.ProductionService, merge *batchfactoryv11.MergeService, automation *FileAutomationStatusReader, version string) {
	if mux == nil || store == nil {
		return
	}
	mux.Handle("GET /api/batch-factory/"+version+"/batches/{batchId}/runtime-summary", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batch, err := runtimeSummaryBatch(r, store, owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		result := map[string]any{
			"batchId":        batch.ID,
			"batchRevision":  batch.Revision,
			"automation":     map[string]any{"state": "idle", "counts": idleAutomationCounts()},
			"production":     map[string]any{"batchId": batch.ID, "jobs": []any{}},
			"merge":          map[string]any{"batchId": batch.ID, "jobs": []any{}},
			"stageSummaries": map[string]any{},
		}
		if automation != nil {
			status, statusErr := automation.Status(owner, batch.ID)
			if statusErr != nil {
				writeStoreError(w, batchfactoryv11.ErrUnavailable)
				return
			}
			result["automation"] = status
		}
		if production != nil {
			status, statusErr := production.GetBatchStatus(r.Context(), owner, batch.ID)
			if statusErr != nil {
				writeStoreError(w, statusErr)
				return
			}
			result["production"] = status
		}
		if merge != nil {
			jobs, statusErr := merge.GetBatchStatus(r.Context(), owner, batch.ID)
			if statusErr != nil {
				writeStoreError(w, statusErr)
				return
			}
			result["merge"] = map[string]any{"batchId": batch.ID, "jobs": jobs}
		}
		if stageService == nil {
			stageService = &batchfactoryv11.BookStageService{Store: store}
		}
		stageSummaries := result["stageSummaries"].(map[string]any)
		for _, book := range batch.Books {
			summary, summaryErr := stageService.Summary(r.Context(), owner, batch.ID, book.ID)
			if summaryErr != nil {
				stageSummaries[book.ID] = map[string]any{"bookId": book.ID, "runs": []any{}, "unavailable": true, "error": summaryErr.Error()}
				continue
			}
			stageSummaries[book.ID] = summary
		}
		writeJSON(w, http.StatusOK, result)
	})))
}

func runtimeSummaryBatch(r *http.Request, store batchfactoryv11.Store, owner, batchID string) (batchfactoryv11.Batch, error) {
	if reader, ok := store.(interface {
		GetBatchRuntimeIndex(context.Context, string, string) (batchfactoryv11.Batch, error)
	}); ok {
		return reader.GetBatchRuntimeIndex(r.Context(), owner, batchID)
	}
	return store.GetBatch(r.Context(), owner, batchID)
}
