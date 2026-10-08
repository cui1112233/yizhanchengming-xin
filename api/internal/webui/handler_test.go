package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
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

	req = httptest.NewRequest(http.MethodGet, "/api/build-info", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() == "" {
		t.Fatalf("build info status=%d body=%q", rec.Code, rec.Body.String())
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

var _ fs.FS
