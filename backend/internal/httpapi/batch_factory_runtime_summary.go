package httpapi

import (
	"context"
	"encoding/json"
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
		comicStore, hasComicPages := store.(batchfactoryv11.ComicPageStore)
		// Stage availability is owner-level. Evaluate it once per request instead
		// of reloading the owner settings for every book in the batch.
		stageAvailability := stageService.Availability(r.Context(), owner)
		for _, book := range batch.Books {
			summary, summaryErr := stageService.SummaryWithAvailability(r.Context(), owner, batch.ID, book.ID, stageAvailability)
			if summaryErr != nil {
				stageSummaries[book.ID] = map[string]any{"bookId": book.ID, "runs": []any{}, "unavailable": true, "error": summaryErr.Error()}
				continue
			}
			// Manual comic-page generation writes durable page images but does not
			// create an automatic stage-run row. Include the page truth in the
			// runtime summary so the UI does not keep saying "waiting to generate"
			// after every page already has a primary image.
			stageSummary := map[string]any{
				"bookId":       summary.BookID,
				"runs":         summary.Runs,
				"availability": summary.Availability,
			}
			if summary.LastFailed != nil {
				stageSummary["lastFailed"] = summary.LastFailed
			}
			if hasComicPages && len(book.Videos) > 0 {
				pages, pagesErr := listComicPagesForView(r.Context(), comicStore, owner, batch.ID, book.ID, true)
				if pagesErr == nil {
					readiness := buildComicReadiness(book.Videos, pages, comicPageGenerationLimit)
					stageSummary["comicImageReadyCount"] = readiness.PrimaryReadyCount
					stageSummary["comicImageTotalCount"] = readiness.PagesTotal
					stageSummary["comicImageReady"] = readiness.ImageReadyCount == readiness.PagesTotal
					stageSummary["comicReadiness"] = readiness
				}
				collections, collectionsErr := comicStore.ListComicCollections(r.Context(), owner, batch.ID, book.ID)
				if collectionsErr == nil {
					collectionCount := 0
					collectionPageCount := 0
					publishedCollectionCount := 0
					for _, collection := range collections {
						if len(collection.Items) == 0 {
							continue
						}
						collectionCount++
						collectionPageCount += len(collection.Items)
						if draft, draftErr := store.GetDraft(r.Context(), owner, collection.ID, n8ComicPublishDraftKind, batch.ID); draftErr == nil {
							var publish struct {
								Status     string `json:"status"`
								CarouselID int64  `json:"carouselId"`
							}
							if json.Unmarshal([]byte(draft.Content), &publish) == nil && publish.Status == "published" && publish.CarouselID > 0 {
								publishedCollectionCount++
							}
						}
					}
					stageSummary["comicCollectionCount"] = collectionCount
					stageSummary["comicCollectionPageCount"] = collectionPageCount
					stageSummary["comicPublishedCollectionCount"] = publishedCollectionCount
					stageSummary["comicCollectionReady"] = collectionCount > 0
				}
			}
			stageSummaries[book.ID] = stageSummary
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
