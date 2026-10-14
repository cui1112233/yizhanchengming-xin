package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalExecutorDownloadRoutesMirrorLegacySemantics(t *testing.T) {
	dir := t.TempDir()
	installer := filepath.Join(dir, "qiantie-executor-v88-0.1.0-win-x64.exe")
	if err := os.WriteFile(installer, []byte("MZ-test"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "local-executor-updates", "stable", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(`{"version":"1.0.3"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterLocalExecutorDownloadRoutes(mux, dir)

	request := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost"+path, nil))
		return rec
	}

	if rec := request("/downloads/local-executor/manifest.json"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "qiantie-executor-v88-0.1.0-win-x64.exe") || !strings.Contains(rec.Body.String(), `"runtime":"go"`) {
		t.Fatalf("manifest status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/qiantie-executor-v88-0.1.0-win-x64.exe"); rec.Code != http.StatusOK || rec.Body.String() != "MZ-test" {
		t.Fatalf("installer status=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/updates/stable/manifest.json"); rec.Code != http.StatusOK || rec.Body.String() != `{"version":"1.0.3"}` {
		t.Fatalf("update manifest status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/missing.exe"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/updates/stable/../../secret.exe"); rec.Code != http.StatusNotFound && rec.Code != http.StatusMovedPermanently {
		t.Fatalf("traversal status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLocalExecutorDownloadKeepsLegacyInstallerAsUnadvertisedRollback(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "yizhan-local-executor-v88-1.0.3-win-x64.exe")
	if err := os.WriteFile(legacy, []byte("legacy"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterLocalExecutorDownloadRoutes(mux, dir)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost/downloads/local-executor/yizhan-local-executor-v88-1.0.3-win-x64.exe", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "legacy" {
		t.Fatalf("legacy status=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestLocalExecutorDownloadMissingInstallerReturns503(t *testing.T) {
	mux := http.NewServeMux()
	RegisterLocalExecutorDownloadRoutes(mux, t.TempDir())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost/downloads/local-executor/qiantie-executor-v88-0.1.0-win-x64.exe", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
