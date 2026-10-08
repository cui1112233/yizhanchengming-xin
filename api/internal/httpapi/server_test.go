package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/webui"
)

func TestErrorRequestIDKeepsMethodAndStaticRouting(t *testing.T) {
	api := NewHandler()
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, httptest.NewRequest("PATCH", "/api/auth/login", nil))
	assertErrorEnvelope(t, rec, 405, "METHOD_NOT_ALLOWED")
	if !strings.Contains(rec.Header().Get("Allow"), "POST") {
		t.Fatalf("missing Allow: %v", rec.Header())
	}
	ui := webui.NewHandler(api, fstest.MapFS{"index.html": {Data: []byte("<h1>SPA</h1>")}}, webui.BuildInfo{})
	rec = httptest.NewRecorder()
	ui.ServeHTTP(rec, httptest.NewRequest("GET", "/workspace/unknown", nil))
	if rec.Code != 200 || rec.Body.String() != "<h1>SPA</h1>" {
		t.Fatalf("static fallback changed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status payload = %q, want %q", body["status"], "ok")
	}
}
