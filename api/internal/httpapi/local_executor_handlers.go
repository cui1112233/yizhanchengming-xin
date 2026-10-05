package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

func (h handler) registerVideoLocalExecutor(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorRegistrationInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	result, err := h.deps.VideoLocalExecutor.Register(r.Context(), input)
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h handler) listVideoLocalExecutors(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	items, err := h.deps.VideoLocalExecutor.List(r.Context())
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"executors": items})
}

func (h handler) getVideoLocalExecutorIdentity(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	identity, err := h.deps.VideoLocalExecutor.Identity(r.Context(), executorBearerToken(r))
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, identity)
}

func (h handler) heartbeatVideoLocalExecutor(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorHeartbeatInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.Heartbeat(r.Context(), executorBearerToken(r), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h handler) completeVideoLocalExecutorTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorCompleteInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.CompleteTask(r.Context(), executorBearerToken(r), r.PathValue("taskId"), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h handler) failVideoLocalExecutorTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorFailInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.FailTask(r.Context(), executorBearerToken(r), r.PathValue("taskId"), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func executorBearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return value
}

func writeLocalExecutorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, video.ErrLocalExecutorUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "local_executor_unauthorized", "message": "local executor credential is invalid"})
	case errors.Is(err, video.ErrLocalExecutorInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "local_executor_invalid", "message": "invalid local executor request"})
	case errors.Is(err, video.ErrLocalExecutorTaskNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "local_executor_task_not_found", "message": "local executor task not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "local_executor_internal", "message": "local executor operation failed"})
	}
}
