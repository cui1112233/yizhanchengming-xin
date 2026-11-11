package httpapi

import (
	"errors"
	"log"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

type bookStageRunInput struct {
	Mode              batchfactoryv11.StageMode      `json:"mode"`
	RequestID         string                         `json:"requestId"`
	VideoID           string                         `json:"videoId,omitempty"`
	TextProvider      *textProviderInput             `json:"textProvider"`
	SmartUnifiedStyle string                         `json:"smartUnifiedStyle"`
	OpeningMeta       batchfactoryv11.PresetSnapshot `json:"openingMeta"`
	Provider          string                         `json:"provider,omitempty"`
	H3                bool                           `json:"h3"`
}

func registerBookStageRoutes(mux *http.ServeMux, service *batchfactoryv11.BookStageService, retryComicPages comicPageRetryFunc) {
	registerBookStageSummaryRoute(mux, service, "v11")
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/stages/{stage}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input bookStageRunInput
		if !decodeJSON(w, r, &input) {
			return
		}
		director, err := directorForRequest(service.Director, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		runner := *service
		runner.Director = director
		runner.SmartUnifiedStyle = input.SmartUnifiedStyle
		runner.OpeningMeta = input.OpeningMeta
		runner.H3Director = input.H3
		runner.ProviderOverride = input.Provider
		summary, err := runner.Run(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), batchfactoryv11.BookStage(r.PathValue("stage")), input.Mode, input.RequestID, input.VideoID)
		if err != nil {
			if errors.Is(err, batchfactoryv11.ErrUnavailable) {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
				return
			}
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, summary)
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/stages/retry", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input bookStageRunInput
		if !decodeJSON(w, r, &input) {
			return
		}
		director, err := directorForRequest(service.Director, input.TextProvider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		runner := *service
		runner.Director = director
		runner.SmartUnifiedStyle = input.SmartUnifiedStyle
		runner.OpeningMeta = input.OpeningMeta
		runner.H3Director = input.H3
		runner.ProviderOverride = input.Provider
		summary, err := runner.RetryLastFailed(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.RequestID, input.VideoID)
		if err != nil {
			// 页级失败（漫画页 generation_failed）不会写阶段运行记录，
			// RetryLastFailed 会判定“没有失败步骤”并返回冲突。此时回退到
			// 页级重试，避免用户点“重试失败步骤”只能看到 409。
			if errors.Is(err, batchfactoryv11.ErrConflict) && retryComicPages != nil {
				retried, retryErr := retryComicPages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
				if retryErr == nil && retried > 0 {
					fallback, summaryErr := service.Summary(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
					if summaryErr != nil {
						writeStoreError(w, summaryErr)
						return
					}
					log.Printf("book stage retry fell back to comic pages batch=%s book=%s pages=%d", r.PathValue("batchId"), r.PathValue("bookId"), retried)
					writeJSON(w, http.StatusCreated, fallback)
					return
				}
			}
			if errors.Is(err, batchfactoryv11.ErrUnavailable) {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
				return
			}
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, summary)
	})
}

// registerBookStageSummaryRoute keeps the read-only stage history available
// even when the executable stages are still gated behind Slice 2.
func registerBookStageSummaryRoute(mux *http.ServeMux, service *batchfactoryv11.BookStageService, version string) {
	mux.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/books/{bookId}/stages", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		summary, err := service.Summary(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, summary)
	})
}
