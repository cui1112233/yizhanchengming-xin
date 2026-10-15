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
		{"/shuihuo-production", http.StatusOK},
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

func TestStaticHandlerServesLegacyNovelFetchAliasAndAdminShell(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":                       &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"admin.html":                       &fstest.MapFile{Data: []byte("<html>admin</html>")},
		"batch-rewrite/index.html":         &fstest.MapFile{Data: []byte("<html>fetch</html>")},
	}
	handler := StaticHandler(assets)
	for _, tc := range []struct {
		path string
		want string
	}{
		{"/batch-rewrite/v2", "fetch"},
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
