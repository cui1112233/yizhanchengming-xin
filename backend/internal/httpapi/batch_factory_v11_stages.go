package httpapi

import (
	"errors"
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

func registerBookStageRoutes(mux *http.ServeMux, service *batchfactoryv11.BookStageService) {
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
