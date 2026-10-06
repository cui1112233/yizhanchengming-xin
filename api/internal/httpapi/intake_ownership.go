package httpapi

import (
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func isElevatedUser(user authn.User) bool {
	role := strings.ToLower(strings.TrimSpace(user.Role))
	return role == "admin" || role == "owner"
}

func (h handler) requireIntakeAccess(pathKey string, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		intakeID, err := parsePositiveID(r.PathValue(pathKey))
		if err != nil {
			h.writeServiceError(w, r, http.StatusBadRequest, "INTAKE_INVALID_REQUEST", "Intake ID 无效", "intake", "authorize", err)
			return
		}
		user, ok := authn.CurrentUser(r.Context())
		if !ok || user.ID <= 0 {
			h.writeServiceError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期", "auth", "authorize_intake", nil)
			return
		}
		if h.deps.IntakeAccess == nil {
			h.writeServiceError(w, r, http.StatusServiceUnavailable, "AUTH_POLICY_UNAVAILABLE", "Intake 权限校验暂不可用", "auth", "authorize_intake", nil)
			return
		}
		allowed, err := h.deps.IntakeAccess.CanAccessIntake(r.Context(), intakeID, user.ID, user.TeamID, isElevatedUser(user))
		if err != nil {
			h.writeServiceError(w, r, http.StatusServiceUnavailable, "AUTH_POLICY_UNAVAILABLE", "Intake 权限校验暂不可用", "auth", "authorize_intake", err)
			return
		}
		if !allowed {
			h.writeServiceError(w, r, http.StatusForbidden, "AUTH_FORBIDDEN", "你没有访问此 Intake 的权限", "auth", "authorize_intake", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
