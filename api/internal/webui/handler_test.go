package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPAHandlerServesStaticAsset(t *testing.T) {
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>home</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	}
	handler, err := newSPAHandler(fs.FS(assets))
	if err != nil { t.Fatal(err) }

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if recorder.Code != http.StatusOK { t.Fatalf("status=%d", recorder.Code) }
	if recorder.Body.String() != "console.log('ok')" { t.Fatalf("body=%q", recorder.Body.String()) }
}

func TestSPAHandlerFallsBackToIndexForApplicationRoute(t *testing.T) {
	assets := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>batch factory</html>")},
	}
	handler, err := newSPAHandler(fs.FS(assets))
	if err != nil { t.Fatal(err) }

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/batch-factory", nil))
	if recorder.Code != http.StatusOK { t.Fatalf("status=%d", recorder.Code) }
	if recorder.Body.String() != "<html>batch factory</html>" { t.Fatalf("body=%q", recorder.Body.String()) }
}

func TestSPAHandlerRejectsMissingIndex(t *testing.T) {
	_, err := newSPAHandler(fstest.MapFS{})
	if err == nil { t.Fatal("expected missing index error") }
}
