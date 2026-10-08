package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

// IntakeAccessChecker is the single object-level policy boundary for intake
// reads and mutations. ListVisibleIntakes must filter in SQL rather than fetch
// a global page and filter it in the handler.
type IntakeAccessChecker interface {
	CanAccessIntake(context.Context, int64, int64, int64, bool) (bool, error)
	ListVisibleIntakes(context.Context, int64, int64, bool) ([]intake.Intake, error)
}

func (h handler) requireIntakeAccess(pathKey string, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		intakeID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || intakeID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "INTAKE_INVALID_REQUEST", "message": "Intake ID 无效"})
			return
		}
		if h.deps.IntakeAccess == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "Intake 权限校验暂不可用"})
			return
		}
		user, ok := authn.CurrentUser(r.Context())
		if !ok || user.ID <= 0 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		allowed, err := h.deps.IntakeAccess.CanAccessIntake(r.Context(), intakeID, user.ID, user.TeamID, intakeElevated(user))
		if err != nil {
			h.writeServiceError(w, r, http.StatusServiceUnavailable, "AUTH_POLICY_UNAVAILABLE", "Intake 权限校验暂不可用", "intake", "access", err)
			return
		}
		if !allowed {
			writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有访问此小说获取批次的权限"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func intakeElevated(user authn.User) bool {
	role := strings.ToLower(strings.TrimSpace(user.Role))
	return role == "admin" || role == "owner"
}
