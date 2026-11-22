package httpapi

import (
	"errors"
	"net/http"

	"qiantie/backend/internal/batchfactoryexecution"
)

type batchExecutionEnsureRequest struct {
	BookIDs         []string       `json:"bookIds"`
	Mode            string         `json:"mode"`
	Target          string         `json:"target"`
	ClientRequestID string         `json:"clientRequestId"`
	Config          map[string]any `json:"config"`
}

func writeBatchExecutionError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, batchfactoryexecution.ErrInvalidCommand):
		status = http.StatusBadRequest
	case errors.Is(err, batchfactoryexecution.ErrNotFound):
		status = http.StatusNotFound
	}
	writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
}

func registerBatchExecutionRoutes(mux *http.ServeMux, auth BridgeAuth, service *batchfactoryexecution.Service) {
	if mux == nil || service == nil {
		return
	}
	mux.Handle("POST /api/batch-factory/v12/batches/{batchId}/executions/ensure", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchExecutionEnsureRequest
		if !decodeJSON(w, req, &input) {
			return
		}
		result, err := service.EnsureExecution(req.Context(), batchfactoryexecution.EnsureCommand{Owner: identity.Username, BatchID: req.PathValue("batchId"), BookIDs: input.BookIDs, Mode: input.Mode, Target: input.Target, ClientRequestID: input.ClientRequestID, Config: input.Config})
		if err != nil {
			writeBatchExecutionError(w, err)
			return
		}
		status := http.StatusOK
		if result.Created {
			status = http.StatusCreated
		}
		writeJSON(w, status, result)
	})))
	mux.Handle("GET /api/batch-factory/v12/executions/{executionId}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		execution, err := service.GetExecution(req.Context(), identity.Username, req.PathValue("executionId"))
		if err != nil {
			writeBatchExecutionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, execution)
	})))
	mux.Handle("GET /api/batch-factory/v12/executions/{executionId}/books/{bookId}/timeline", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		events, err := service.Timeline(req.Context(), identity.Username, req.PathValue("executionId"), req.PathValue("bookId"))
		if err != nil {
			writeBatchExecutionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	})))
	for _, action := range []string{"reconcile", "cancel"} {
		action := action
		mux.Handle("POST /api/batch-factory/v12/executions/{executionId}/"+action, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			var execution batchfactoryexecution.Execution
			var err error
			if action == "cancel" {
				execution, err = service.Cancel(req.Context(), identity.Username, req.PathValue("executionId"))
			} else {
				execution, err = service.Reconcile(req.Context(), identity.Username, req.PathValue("executionId"))
			}
			if err != nil {
				writeBatchExecutionError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, execution)
		})))
	}
}
