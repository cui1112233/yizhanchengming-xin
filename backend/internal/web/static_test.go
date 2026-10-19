package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"testing/fstest"
)

func TestStaticHandlerServesIndexAndAssets(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":                   &fstest.MapFile{Data: []byte("<html>home</html>")},
		"assets/app.js":                &fstest.MapFile{Data: []byte("console.log('ok')")},
		"assets/style.css":             &fstest.MapFile{Data: []byte("body{}")},
		"pets/stacky/spritesheet.webp": &fstest.MapFile{Data: []byte("pet")},
	}
	handler := StaticHandler(assets)

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/shuihuo-production", http.StatusFound},
		{"/assets/app.js", http.StatusOK},
		{"/pets/stacky/spritesheet.webp", http.StatusOK},
		{"/missing.js", http.StatusNotFound},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != tc.want {
			t.Fatalf("%s status = %d, want %d", tc.path, res.Code, tc.want)
		}
		if tc.path == "/shuihuo-production" && res.Header().Get("Location") != "/batch-factory" {
			t.Fatalf("%s Location = %q, want /batch-factory", tc.path, res.Header().Get("Location"))
		}
		if tc.want == http.StatusOK && res.Header().Get("Cache-Control") != "no-store, must-revalidate" {
			t.Fatalf("%s Cache-Control = %q, want no-store", tc.path, res.Header().Get("Cache-Control"))
		}
	}
}

func TestStaticHandlerRejectsNilFilesystem(t *testing.T) {
	var empty fs.FS
	handler := StaticHandler(empty)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}
}

func TestStaticHandlerServesAdminShell(t *testing.T) {
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"admin.html": &fstest.MapFile{Data: []byte("<html>admin</html>")},
	}
	handler := StaticHandler(assets)
	for _, tc := range []struct {
		path string
		want string
	}{
		{"/admin", "admin"},
		{"/admin/presets", "admin"},
		{"/novel-panel", "shell"},
	} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), tc.want) {
			t.Fatalf("%s returned status %d body %q, want %q", tc.path, res.Code, res.Body.String(), tc.want)
		}
	}
}

func TestStaticHandlerServesReactNovelFetchOnly(t *testing.T) {
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>react-shell</html>")},
	}
	for _, route := range []string{"/novel-fetch", "/novel-fetch-react"} {
		res := httptest.NewRecorder()
		StaticHandler(assets).ServeHTTP(res, httptest.NewRequest(http.MethodGet, route, nil))
		if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "react-shell") {
			t.Fatalf("react novel-fetch route %s status = %d body %q", route, res.Code, res.Body.String())
		}
	}
	res := httptest.NewRecorder()
	StaticHandler(assets).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/batch-rewrite/v2", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("removed legacy novel-fetch route status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func TestStaticHandlerRejectsRemovedNovelPanelWorkbench(t *testing.T) {
	assets := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")}}
	res := httptest.NewRecorder()
	StaticHandler(assets).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/novel-panel/workbench/index.html", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("removed workbench status = %d, want %d", res.Code, http.StatusNotFound)
	}
}
