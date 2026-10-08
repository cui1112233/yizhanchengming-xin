package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type RuntimeService interface {
	RetryBookRun(context.Context, int64) (task9runtime.WorkItem, bool, error)
	ProjectIDForBookRun(context.Context, int64) (int64, error)
}

type runtimeHTTP struct {
	deps    Dependencies
	runtime RuntimeService
}

// NewHandlerWithRuntime adds the Task9 browser mutation on top of the existing
// API router while reusing the exact Task15 same-origin, auth, capability and
// ownership boundaries. All other routes continue through NewHandler.
func NewHandlerWithRuntime(deps Dependencies, runtime RuntimeService) http.Handler {
	base := NewHandler(deps)
	if runtime == nil {
		return base
	}
	rh := runtimeHTTP{deps: deps, runtime: runtime}
	boundary := handler{deps: deps}
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/batch-projects/{projectId}/book-runs/{bookRunId}/retry",
		boundary.withObservability(boundary.requireSameOrigin(boundary.requireCapability(CapabilityBatchExecute, http.HandlerFunc(rh.retryBookRun)))))
	mux.Handle("/", base)
	return mux
}

func (h runtimeHTTP) retryBookRun(w http.ResponseWriter, r *http.Request) {
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookRunID, err := parsePositiveID(r.PathValue("bookRunId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	actualProjectID, err := h.runtime.ProjectIDForBookRun(r.Context(), bookRunID)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "RUNTIME_NOT_FOUND", "message": "BookRun 不存在"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "RUNTIME_UNAVAILABLE", "message": "Runtime 暂不可用"})
		return
	}
	if actualProjectID != projectID {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "RUNTIME_NOT_FOUND", "message": "BookRun 不属于该项目"})
		return
	}
	boundary := handler{deps: h.deps}
	if !boundary.authorizeBatchProject(w, r, projectID) {
		return
	}

	item, created, err := h.runtime.RetryBookRun(r.Context(), bookRunID)
	if errors.Is(err, task9runtime.ErrBookRunNotRetryable) {
		writeJSON(w, http.StatusConflict, map[string]any{"code": "RUNTIME_NOT_RETRYABLE", "message": "该 BookRun 当前不可重试"})
		return
	}
	if errors.Is(err, task9runtime.ErrRuntimeQueueUnavailable) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "RUNTIME_QUEUE_UNAVAILABLE", "message": "Runtime Queue 暂不可用"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": "RUNTIME_RETRY_FAILED", "message": "BookRun 重试失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookRunId": item.BookRunID, "bookId": item.BookID, "attempt": item.Attempt, "created": created})
}
