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
	windowsInstaller := filepath.Join(dir, "一战晟铭执行器-安装版.exe")
	if err := os.WriteFile(windowsInstaller, []byte("PK-test"), 0o600); err != nil {
		t.Fatal(err)
	}
	macInstaller := filepath.Join(dir, "doubao-relay_1.0.0_universal.dmg")
	if err := os.WriteFile(macInstaller, []byte("dmg-test"), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterLocalExecutorDownloadRoutes(mux, dir)

	request := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost"+path, nil))
		return rec
	}

	if rec := request("/downloads/local-executor/manifest.json"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "一战晟铭执行器-安装版.exe") || !strings.Contains(rec.Body.String(), `"windows":"1.2.0"`) || !strings.Contains(rec.Body.String(), `"macos":"1.0.0"`) || strings.Contains(rec.Body.String(), "windows_x64.zip") || !strings.Contains(rec.Body.String(), "doubao-relay_1.0.0_universal.dmg") || !strings.Contains(rec.Body.String(), `"runtime":"doubao_relay"`) {
		t.Fatalf("manifest status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/一战晟铭执行器-安装版.exe"); rec.Code != http.StatusOK || rec.Body.String() != "PK-test" || rec.Header().Get("Content-Type") != "application/vnd.microsoft.portable-executable" {
		t.Fatalf("windows installer status=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec := request("/downloads/local-executor/doubao-relay_1.0.0_universal.dmg"); rec.Code != http.StatusOK || rec.Body.String() != "dmg-test" {
		t.Fatalf("mac installer status=%d body=%q", rec.Code, rec.Body.String())
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
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost/downloads/local-executor/一战晟铭执行器-安装版.exe", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
