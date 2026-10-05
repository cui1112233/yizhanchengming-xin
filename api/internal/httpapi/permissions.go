package httpapi

import (
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

const (
	CapabilityBatchView                = "batch.view"
	CapabilityBatchExecute             = "batch.execute"
	CapabilityBatchConfigure           = "batch.configure"
	CapabilityVideoProviderConfigure   = "video.provider.configure"
	CapabilityPublishConfigure         = "publish.configure"
	CapabilityPublishAccountConfigure  = "publish.account.configure"
	CapabilityPublishExecute           = "publish.execute"
	CapabilityPublishAuditView         = "publish.audit.view"
)

func userHasCapability(user authn.User, capability string) bool {
	role := strings.ToLower(strings.TrimSpace(user.Role))
	if role == "admin" || role == "owner" {
		return true
	}
	for _, item := range user.Capabilities {
		if item == capability {
			return true
		}
	}
	return false
}

func (h handler) requireCapability(capability string, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authn.CurrentUser(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		required := capability
		// Task 14 originally reused batch.configure for provider configuration.
		// Provider credentials are a separate secret-management boundary, so keep
		// the existing route wiring but translate that route to its dedicated
		// capability. Admin/owner retain their existing elevated semantics.
		if capability == CapabilityBatchConfigure && strings.HasPrefix(r.URL.Path, "/api/v1/video-providers/") {
			required = CapabilityVideoProviderConfigure
		}
		if !userHasCapability(user, required) {
			writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有执行此操作的权限"})
			return
		}
		next.ServeHTTP(w, r)
	}))
}
