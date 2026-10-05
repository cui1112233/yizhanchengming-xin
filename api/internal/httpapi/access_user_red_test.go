package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

func TestAuthenticatedAccessLogIncludesStableUserID(t *testing.T) {
	var logs bytes.Buffer
	auth := opsAuthService{user: authn.User{ID: 42, Role: "admin"}}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/current-user", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Auth: auth, Logger: observability.NewJSONLogger(&logs)}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(logs.String(), `"user_id":42`) {
		t.Fatalf("authenticated access log missing stable user id: %s", logs.String())
	}
}
