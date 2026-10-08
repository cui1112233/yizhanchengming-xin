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

var _ fs.FS
