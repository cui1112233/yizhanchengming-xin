package webui

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandlerServesEmbeddedAssetsAndSPAEntry(t *testing.T) {
	ui := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("<main>workspace</main>")},
		"assets/app-abc123.js": &fstest.MapFile{Data: []byte("console.log('workspace')")},
	}
	h := NewHandler(http.NotFoundHandler(), ui, BuildInfo{GitSHA: "abc123"})

	for _, path := range []string{"/", "/history", "/settings"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d want 200", path, rec.Code)
		}
		if got := rec.Header().Get("X-YCM-Static-Source"); got != "go-embed" {
			t.Fatalf("%s static source=%q want go-embed", path, got)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/app-abc123.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log('workspace')" {
		t.Fatalf("asset status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-YCM-Static-Source"); got != "go-embed" {
		t.Fatalf("asset static source=%q want go-embed", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/build-info", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("build info status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("build info content type=%q", got)
	}
	var info BuildInfo
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode build info: %v", err)
	}
	if info.GitSHA != "abc123" {
		t.Fatalf("build info gitSha=%q want abc123", info.GitSHA)
	}
}

func TestHandlerLeavesAPIRoutesWithAPIServer(t *testing.T) {
	called := false
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/api/auth/current-user" {
			t.Fatalf("api path=%q", r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
	h := NewHandler(api, fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}, BuildInfo{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/current-user", nil))
	if !called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("called=%v status=%d", called, rec.Code)
	}
}

func TestHandlerServesAdminSPASeparatelyFromUserSPA(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
	userUI := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<main>user</main>")},
	}
	adminUI := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("<main>admin</main>")},
		"assets/admin-test.js": &fstest.MapFile{Data: []byte("console.log('admin')")},
	}
	h := NewHandlerWithAdmin(api, userUI, adminUI, BuildInfo{GitSHA: "abc123"})

	for _, requestPath := range []string{"/admin", "/admin/", "/admin/prompts"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, requestPath, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != "<main>admin</main>" {
			t.Fatalf("%s status=%d body=%q", requestPath, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("X-YCM-Static-Source"); got != "go-embed" {
			t.Fatalf("%s static source=%q", requestPath, got)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/assets/admin-test.js", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log('admin')" {
		t.Fatalf("admin asset status=%d body=%q", rec.Code, rec.Body.String())
	}

	called := false
	api = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/api/v1/admin/prompts" {
			t.Fatalf("api path=%q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h = NewHandlerWithAdmin(api, userUI, adminUI, BuildInfo{})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/prompts", nil))
	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("api called=%v status=%d", called, rec.Code)
	}
}

func TestHandlerDeepRefreshAndAssetsStayInTheirEmbeddedApplication(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "api-only", http.StatusTeapot)
	})
	userUI := fstest.MapFS{
		"index.html":          &fstest.MapFile{Data: []byte("<main>user-spa</main>")},
		"assets/user-test.js": &fstest.MapFile{Data: []byte("user-asset")},
	}
	adminUI := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("<main>admin-spa</main>")},
		"assets/admin-test.js": &fstest.MapFile{Data: []byte("admin-asset")},
	}
	h := NewHandlerWithAdmin(api, userUI, adminUI, BuildInfo{})

	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "user deep refresh", path: "/history/deep-link", body: "<main>user-spa</main>"},
		{name: "admin deep refresh", path: "/admin/prompts/deep-link", body: "<main>admin-spa</main>"},
		{name: "user asset", path: "/assets/user-test.js", body: "user-asset"},
		{name: "admin asset", path: "/admin/assets/admin-test.js", body: "admin-asset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != http.StatusOK || rec.Body.String() != tt.body {
				t.Fatalf("status=%d body=%q want 200 %q", rec.Code, rec.Body.String(), tt.body)
			}
			if got := rec.Header().Get("X-YCM-Static-Source"); got != "go-embed" {
				t.Fatalf("static source=%q want go-embed", got)
			}
		})
	}
}

func TestHandlerNeverFallsBackToSPAForAPIOrNonReadRequests(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"owner":"api"}`))
	})
	h := NewHandlerWithAdmin(
		api,
		fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("user-spa")}},
		fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("admin-spa")}},
		BuildInfo{},
	)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api"},
		{method: http.MethodGet, path: "/api/unknown"},
		{method: http.MethodGet, path: "/api/v1/admin/unknown"},
		{method: http.MethodPost, path: "/history/deep-link"},
		{method: http.MethodDelete, path: "/admin/prompts/deep-link"},
		{method: http.MethodPost, path: "/api/build-info"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != http.StatusTeapot || rec.Body.String() != `{"owner":"api"}` {
				t.Fatalf("status=%d body=%q want API response", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), "user-spa") || strings.Contains(rec.Body.String(), "admin-spa") {
				t.Fatalf("request received SPA fallback: %q", rec.Body.String())
			}
			if got := rec.Header().Get("X-YCM-Static-Source"); got != "" {
				t.Fatalf("API response static source=%q want empty", got)
			}
		})
	}
}

var _ fs.FS
