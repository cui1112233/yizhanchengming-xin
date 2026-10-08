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

type resolvedBatchProjectContextKey struct{}

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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), resolvedBatchProjectContextKey{}, projectID)))
	})
}

func (h handler) requireActiveResolvedBatchProject(next http.Handler) http.Handler {
	if h.deps.Auth == nil && h.deps.BatchProjectLifecycle == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID, ok := r.Context().Value(resolvedBatchProjectContextKey{}).(int64)
		if !ok || projectID <= 0 {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目状态校验暂不可用"})
			return
		}
		if h.ensureBatchProjectActive(w, r, projectID) {
			next.ServeHTTP(w, r)
		}
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

type batchProjectAccessDecision uint8

const (
	batchProjectAccessAllowed batchProjectAccessDecision = iota
	batchProjectAccessUnauthenticated
	batchProjectAccessForbidden
	batchProjectAccessUnavailable
)

func (h handler) requireVideoResourceAccess(pathKey string, resolve videoProjectResolverFunc, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		if h.deps.BatchProjectLifecycle == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resourceID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
			if err != nil || resourceID <= 0 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"code": "VIDEO_INVALID_REQUEST", "message": "VIDEO resource ID 无效"})
				return
			}
			projectID, err := resolve(r.Context(), resourceID)
			if errors.Is(err, video.ErrNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]any{"code": "VIDEO_NOT_FOUND", "message": "VIDEO resource 不存在"})
				return
			}
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目状态校验暂不可用"})
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), resolvedBatchProjectContextKey{}, projectID)))
		})
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
		switch h.evaluateBatchProjectAccess(r, projectID) {
		case batchProjectAccessAllowed:
			// Continue to the business handler.
		case batchProjectAccessUnauthenticated:
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		case batchProjectAccessForbidden:
			// Resource-only URLs intentionally collapse a foreign project into the
			// same envelope as an absent resource to avoid an ownership oracle.
			writeJSON(w, http.StatusNotFound, map[string]any{"code": "VIDEO_NOT_FOUND", "message": "VIDEO resource 不存在"})
			return
		default:
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), resolvedBatchProjectContextKey{}, projectID)))
	})
}

func (h handler) authorizeBatchProject(w http.ResponseWriter, r *http.Request, projectID int64) bool {
	switch h.evaluateBatchProjectAccess(r, projectID) {
	case batchProjectAccessAllowed:
		return true
	case batchProjectAccessUnavailable:
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
	case batchProjectAccessUnauthenticated:
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
	case batchProjectAccessForbidden:
		writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有访问此批量项目的权限"})
	}
	return false
}

// evaluateBatchProjectAccess has no response side effects. Direct project URLs
// map forbidden to 403, while indirect resource URLs may safely collapse it to
// their domain-specific 404 envelope.
func (h handler) evaluateBatchProjectAccess(r *http.Request, projectID int64) batchProjectAccessDecision {
	if h.deps.Auth == nil {
		return batchProjectAccessAllowed
	}
	if h.deps.BatchProjectAccess == nil {
		return batchProjectAccessUnavailable
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		return batchProjectAccessUnauthenticated
	}
	role := strings.ToLower(strings.TrimSpace(user.Role))
	elevated := role == "admin" || role == "owner"
	allowed, err := h.deps.BatchProjectAccess.CanAccessBatchProject(r.Context(), projectID, user.ID, user.TeamID, elevated)
	if err != nil {
		return batchProjectAccessUnavailable
	}
	if !allowed {
		return batchProjectAccessForbidden
	}
	return batchProjectAccessAllowed
}
