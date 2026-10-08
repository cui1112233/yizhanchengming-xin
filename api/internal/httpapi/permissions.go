package httpapi

import (
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

const (
	CapabilityBatchView               = "batch.view"
	CapabilityBatchExecute            = "batch.execute"
	CapabilityBatchConfigure          = "batch.configure"
	CapabilityPublishConfigure        = "publish.configure"
	CapabilityPublishAccountConfigure = "publish.account.configure"
	CapabilityPublishExecute          = "publish.execute"
	CapabilityPublishAuditView        = "publish.audit.view"
)

func userHasCapability(user authn.User, capability string) bool {
	return authn.HasCapability(user.Capabilities, capability)
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
		if !userHasCapability(user, capability) {
			writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有执行此操作的权限"})
			return
		}
		next.ServeHTTP(w, r)
	}))
}
