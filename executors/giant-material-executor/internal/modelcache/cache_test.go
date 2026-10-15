package modelcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEnsureReusesCompleteCacheWithoutNetwork(t *testing.T) {
	root := t.TempDir()
	payload := []byte("cached-ocr-runtime")
	manifest := testManifest("v1", payload, "http://127.0.0.1:1/unused")
	writeActive(t, root, manifest, payload)

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()
	manifest.PackageURL = server.URL + "/runtime.bin"

	install, err := (Cache{Root: root}).Ensure(context.Background(), manifest, nil)
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 0 {
		t.Fatalf("expected no network requests, got %d", requests.Load())
	}
	if install.Version != manifest.Version || install.Dir != filepath.Join(root, manifest.Version) {
		t.Fatalf("unexpected install: %+v", install)
	}
}

func TestEnsureResumesInterruptedDownloadWithRange(t *testing.T) {
	root := t.TempDir()
	payload := []byte("resumable-ocr-runtime")
	manifest := testManifest("v2", payload, "")
	partial := payload[:7]
	if err := os.WriteFile(filepath.Join(root, manifest.Version+".part"), partial, 0o600); err != nil {
		t.Fatal(err)
	}

	var seenRange atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRange.Store(r.Header.Get("Range"))
		if r.Header.Get("Range") != "bytes=7-" {
			http.Error(w, "range required", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)-len(partial)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload[len(partial):])
	}))
	defer server.Close()
	manifest.PackageURL = server.URL + "/runtime.bin"

	if _, err := (Cache{Root: root}).Ensure(context.Background(), manifest, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := seenRange.Load().(string); got != "bytes=7-" {
		t.Fatalf("range=%q", got)
	}
	assertFileBytes(t, filepath.Join(root, manifest.Version, bundleName), payload)
}

func TestEnsureHashFailureLeavesExistingActiveVersionUntouched(t *testing.T) {
	root := t.TempDir()
	oldPayload := []byte("old-runtime")
	newPayload := []byte("corrupted-runtime")
	manifest := testManifest("v3", []byte("expected-runtime!"), "")
	writeActive(t, root, Manifest{Version: manifest.Version, Size: int64(len(oldPayload)), SHA256: sha256Hex(oldPayload)}, oldPayload)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(newPayload)
	}))
	defer server.Close()
	manifest.PackageURL = server.URL + "/runtime.bin"

	if _, err := (Cache{Root: root}).Ensure(context.Background(), manifest, nil); err == nil {
		t.Fatal("expected hash mismatch")
	} else if !strings.Contains(err.Error(), "MODEL_HASH_MISMATCH") {
		t.Fatalf("error=%v", err)
	}
	assertFileBytes(t, filepath.Join(root, manifest.Version, bundleName), oldPayload)
}

func TestEnsureReportsCompletionOnlyAfterAtomicActivation(t *testing.T) {
	root := t.TempDir()
	payload := []byte("atomic-runtime")
	manifest := testManifest("v4", payload, "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	manifest.PackageURL = server.URL + "/runtime.bin"

	var completedBeforeActivation bool
	progress := func(update DownloadProgress) {
		if update.Percent == 100 {
			if _, err := os.Stat(filepath.Join(root, manifest.Version, bundleName)); err != nil {
				completedBeforeActivation = true
			}
		}
	}
	if _, err := (Cache{Root: root}).Ensure(context.Background(), manifest, progress); err != nil {
		t.Fatal(err)
	}
	if completedBeforeActivation {
		t.Fatal("completion reported before active directory was available")
	}
}

func testManifest(version string, payload []byte, packageURL string) Manifest {
	return Manifest{Version: version, PackageURL: packageURL, Size: int64(len(payload)), SHA256: sha256Hex(payload), RuntimeVersion: "python-paddleocr-1"}
}

func writeActive(t *testing.T, root string, manifest Manifest, payload []byte) {
	t.Helper()
	dir := filepath.Join(root, manifest.Version)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	assertErr(t, os.WriteFile(filepath.Join(dir, bundleName), payload, 0o600))
	data, err := json.Marshal(manifest)
	assertErr(t, err)
	assertErr(t, os.WriteFile(filepath.Join(dir, manifestName), append(data, '\n'), 0o600))
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s=%q, want %q", path, got, want)
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
