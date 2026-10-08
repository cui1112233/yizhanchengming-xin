package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type BatchProjectAccessChecker interface {
	CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error)
}

type VideoResourceProjectResolver interface {
	ProjectIDForProductionTask(context.Context, int64) (int64, error)
	ProjectIDForMergeJob(context.Context, int64) (int64, error)
	ProjectIDForMergeAttempt(context.Context, int64) (int64, error)
}

func (h handler) requireBatchProjectAccess(pathKey string, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || projectID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "BATCH_PROJECT_INVALID_REQUEST", "message": "批量项目 ID 无效"})
			return
		}
		if !h.authorizeBatchProject(w, r, projectID) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h handler) requireVideoTaskAccess(next http.Handler) http.Handler {
	return h.requireVideoResourceAccess("taskId", func(ctx context.Context, id int64) (int64, error) {
		if h.deps.VideoResourceProjects == nil {
			return 0, errVideoAccessUnavailable
		}
		return h.deps.VideoResourceProjects.ProjectIDForProductionTask(ctx, id)
	}, next)
}

func (h handler) requireVideoMergeJobAccess(next http.Handler) http.Handler {
	return h.requireVideoResourceAccess("jobId", func(ctx context.Context, id int64) (int64, error) {
		if h.deps.VideoResourceProjects == nil {
			return 0, errVideoAccessUnavailable
		}
		return h.deps.VideoResourceProjects.ProjectIDForMergeJob(ctx, id)
	}, next)
}

func (h handler) requireVideoMergeAttemptAccess(next http.Handler) http.Handler {
	return h.requireVideoResourceAccess("attemptId", func(ctx context.Context, id int64) (int64, error) {
		if h.deps.VideoResourceProjects == nil {
			return 0, errVideoAccessUnavailable
		}
		return h.deps.VideoResourceProjects.ProjectIDForMergeAttempt(ctx, id)
	}, next)
}

var errVideoAccessUnavailable = errors.New("video access policy unavailable")

type videoProjectResolverFunc func(context.Context, int64) (int64, error)

func (h handler) requireVideoResourceAccess(pathKey string, resolve videoProjectResolverFunc, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resourceID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || resourceID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "VIDEO_INVALID_REQUEST", "message": "VIDEO resource ID 无效"})
			return
		}
		if h.deps.BatchProjectAccess == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
			return
		}
		projectID, err := resolve(r.Context(), resourceID)
		if errors.Is(err, video.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"code": "VIDEO_NOT_FOUND", "message": "VIDEO resource 不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
			return
		}
		if !h.authorizeBatchProject(w, r, projectID) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h handler) authorizeBatchProject(w http.ResponseWriter, r *http.Request, projectID int64) bool {
	if h.deps.Auth == nil {
		return true
	}
	if h.deps.BatchProjectAccess == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
		return false
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return false
	}
	role := strings.ToLower(strings.TrimSpace(user.Role))
	elevated := role == "admin" || role == "owner"
	allowed, err := h.deps.BatchProjectAccess.CanAccessBatchProject(r.Context(), projectID, user.ID, user.TeamID, elevated)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
		return false
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有访问此批量项目的权限"})
		return false
	}
	return true
}
