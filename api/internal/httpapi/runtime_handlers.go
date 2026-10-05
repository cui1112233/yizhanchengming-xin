package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type RuntimeService interface {
	RetryBookRun(context.Context, int64) (task9runtime.WorkItem, bool, error)
	ProjectIDForBookRun(context.Context, int64) (int64, error)
}

func (h handler) retryRuntimeBookRun(w http.ResponseWriter, r *http.Request) {
	if h.deps.Runtime == nil {
		writeError(w, http.StatusServiceUnavailable, "runtime service unavailable")
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	bookRunID, err := parsePositiveID(r.PathValue("bookRunId"))
	if err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }

	actualProjectID, err := h.deps.Runtime.ProjectIDForBookRun(r.Context(), bookRunID)
	if errors.Is(err, sql.ErrNoRows) { writeJSON(w,http.StatusNotFound,map[string]any{"code":"RUNTIME_NOT_FOUND","message":"BookRun 不存在"}); return }
	if err != nil { writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"RUNTIME_UNAVAILABLE","message":"Runtime 暂不可用"}); return }
	if actualProjectID != projectID { writeJSON(w,http.StatusNotFound,map[string]any{"code":"RUNTIME_NOT_FOUND","message":"BookRun 不属于该项目"}); return }
	if !h.authorizeRuntimeProject(w,r,projectID) { return }

	item, created, err := h.deps.Runtime.RetryBookRun(r.Context(), bookRunID)
	if errors.Is(err, task9runtime.ErrBookRunNotRetryable) { writeJSON(w,http.StatusConflict,map[string]any{"code":"RUNTIME_NOT_RETRYABLE","message":"该 BookRun 当前不可重试"}); return }
	if errors.Is(err, task9runtime.ErrRuntimeQueueUnavailable) { writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"RUNTIME_QUEUE_UNAVAILABLE","message":"Runtime Queue 暂不可用"}); return }
	if err != nil { writeJSON(w,http.StatusInternalServerError,map[string]any{"code":"RUNTIME_RETRY_FAILED","message":"BookRun 重试失败"}); return }
	writeJSON(w,http.StatusOK,map[string]any{"bookRunId":item.BookRunID,"bookId":item.BookID,"attempt":item.Attempt,"created":created})
}

func (h handler) authorizeRuntimeProject(w http.ResponseWriter, r *http.Request, projectID int64) bool {
	if h.deps.Auth == nil { return true }
	if h.deps.BatchProjectAccess == nil { writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"AUTH_POLICY_UNAVAILABLE","message":"项目权限校验暂不可用"}); return false }
	user, ok := authn.CurrentUser(r.Context()); if !ok || user.ID<=0 { writeJSON(w,http.StatusUnauthorized,map[string]any{"code":"AUTH_UNAUTHENTICATED","message":"登录状态无效或已过期"}); return false }
	role:=strings.ToLower(strings.TrimSpace(user.Role)); elevated:=role=="admin"||role=="owner"
	allowed,err:=h.deps.BatchProjectAccess.CanAccessBatchProject(r.Context(),projectID,user.ID,user.TeamID,elevated)
	if err!=nil { writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"AUTH_POLICY_UNAVAILABLE","message":"项目权限校验暂不可用"}); return false }
	if !allowed { writeJSON(w,http.StatusForbidden,map[string]any{"code":"AUTH_FORBIDDEN","message":"你没有访问此批量项目的权限"}); return false }
	return true
}
