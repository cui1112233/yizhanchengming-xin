package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestSecurityBatchConfigureDoesNotGrantProviderSecretConfigure(t *testing.T) {
	auth := task14AuthStub{user: authn.User{
		ID:           7,
		Role:         "member",
		Capabilities: []string{CapabilityBatchConfigure},
	}}
	h := NewHandler(Dependencies{Auth: auth, VideoConfig: phase2HTTPConfigService{}})
	req := httptest.NewRequest(http.MethodPut, "http://app.example/api/v1/video-providers/personal_api/models/yd2.0-mini", strings.NewReader(`{"secret":"provider-secret","enabled":true}`))
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s, want 403: batch.configure must not grant provider secret configuration", rec.Code, rec.Body.String())
	}
}
