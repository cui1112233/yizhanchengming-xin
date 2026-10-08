package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestAdminCapabilityMiddlewareUsesEffectiveCapabilitiesNotRole(t *testing.T) {
	for _, tc := range []struct {
		name, role   string
		capabilities []string
		want         int
	}{
		{"legacy admin", "admin", authn.EffectiveCapabilities("admin", nil), http.StatusNoContent},
		{"owner", "owner", authn.EffectiveCapabilities("owner", nil), http.StatusNoContent},
		{"dev", "dev", authn.EffectiveCapabilities("dev", nil), http.StatusNoContent},
		{"member", "member", nil, http.StatusForbidden},
		{"view only manager reads", "manager", []string{authn.CapabilityAdminPromptView}, http.StatusNoContent},
		{"view only manager cannot publish", "manager", []string{authn.CapabilityAdminPromptView}, http.StatusForbidden},
		{"role alone cannot bypass", "owner", nil, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			required := authn.CapabilityAdminPromptView
			if tc.name == "view only manager cannot publish" {
				required = authn.CapabilityAdminPromptPublish
			}
			h := handler{deps: Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Role: tc.role, Capabilities: tc.capabilities}}}}
			protected := h.requireCapability(required, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prompts", nil)
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}
