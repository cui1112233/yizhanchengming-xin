package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"testing/fstest"
)

func TestStaticHandlerServesIndexAndAssets(t *testing.T) {
	assets := fstest.MapFS{
		"index.html":       &fstest.MapFile{Data: []byte("<html>home</html>")},
		"assets/app.js":    &fstest.MapFile{Data: []byte("console.log('ok')")},
		"assets/style.css": &fstest.MapFile{Data: []byte("body{}")},
	}
	handler := StaticHandler(assets)

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/shuihuo-production", http.StatusOK},
		{"/assets/app.js", http.StatusOK},
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
