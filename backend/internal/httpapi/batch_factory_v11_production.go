package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

type productionSubmitInput struct {
	RequestID     string `json:"requestId"`
	Provider      string `json:"provider,omitempty"`
	CompilationID string `json:"compilationId,omitempty"`
}

type videoProviderConfigInput struct {
	Provider   string `json:"provider"`
	APIKey     string `json:"apiKey,omitempty"`
	Model      string `json:"model,omitempty"`
	CreateURL  string `json:"createUrl,omitempty"`
	TasksURL   string `json:"tasksUrl,omitempty"`
	ResultURL  string `json:"resultUrl,omitempty"`
	WorkflowID string `json:"workflowId,omitempty"`
}

func registerProductionRoutes(mux *http.ServeMux, service *batchfactoryv11.ProductionService, catalogStores ...novelfetchworkshop.Store) {
	registerProductionStatusRoute(mux, service, "v11")
	mux.HandleFunc("PUT /api/batch-factory/v11/video-provider/config", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if service == nil || service.ProviderRegistry == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		var input videoProviderConfigInput
		if !decodeJSON(w, r, &input) {
			return
		}
		err := service.ProviderRegistry.Put(r.Context(), owner, batchfactoryv11.VideoProviderConfig{
			Provider: input.Provider, APIKey: input.APIKey, Model: input.Model,
			CreateURL: input.CreateURL, TasksURL: input.TasksURL, ResultURL: input.ResultURL, WorkflowID: input.WorkflowID,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		provider := batchfactoryv11.NormalizeVideoProviderForHTTP(input.Provider)
		model := input.Model
		if provider == batchfactoryv11.VideoProviderPersonalAPI && strings.TrimSpace(model) == "" {
			model = batchfactoryv11.DefaultPersonalVideoModel
		}
		writeJSON(w, http.StatusOK, map[string]any{"provider": provider, "model": model, "configured": true})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/video-provider/status", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		provider := batchfactoryv11.NormalizeVideoProviderForHTTP(r.URL.Query().Get("provider"))
		if service == nil || service.ProviderRegistry == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		view, err := service.ProviderRegistry.View(r.Context(), owner, provider)
		if err != nil && !errors.Is(err, batchfactoryv11.ErrUnavailable) {
			writeStoreError(w, err)
			return
		}
		if errors.Is(err, batchfactoryv11.ErrUnavailable) {
			view = batchfactoryv11.VideoProviderConfigView{Provider: provider, Configured: false}
		}
		writeJSON(w, http.StatusOK, view)
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/production", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input productionSubmitInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.RequestID) == "" {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		if service != nil && service.ProviderRegistry != nil {
			if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), service.Store, service.ProviderRegistry, owner, catalogStores...); syncErr != nil {
				writeStoreError(w, syncErr)
				return
			}
		}
		job, err := service.SubmitBookProductionWithOptions(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input.RequestID, input.Provider, batchfactoryv11.ProductionOptions{CompilationID: input.CompilationID})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"job": job})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/production", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input productionSubmitInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.RequestID) == "" {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		if service != nil && service.ProviderRegistry != nil {
			if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), service.Store, service.ProviderRegistry, owner, catalogStores...); syncErr != nil {
				writeStoreError(w, syncErr)
				return
			}
		}
		status, err := service.SubmitBatchProductionWithProvider(r.Context(), owner, r.PathValue("batchId"), input.RequestID, input.Provider)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, status)
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/production/cancel", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if service == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		var input productionSubmitInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.RequestID) == "" {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		result, err := service.CancelBatch(r.Context(), owner, r.PathValue("batchId"), input.RequestID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	mux.HandleFunc("DELETE /api/batch-factory/v11/batches/{batchId}/books/{bookId}/videos/{videoId}/tasks/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if service == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		if err := service.RemoveBookProductionCandidate(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), r.PathValue("taskId")); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	/* status is registered by registerProductionStatusRoute above */
}

func registerProductionStatusRoute(mux *http.ServeMux, service *batchfactoryv11.ProductionService, version string) {
	mux.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/status", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		status, err := service.GetBatchStatus(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
}
