package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeReleaseManifest(t *testing.T, releaseID string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "runtime-release.json")
	contents, err := json.Marshal(map[string]string{"releaseId": releaseID, "gitSha": "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProductUpdatesLatestReturnsNotesForCurrentRelease(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: writeReleaseManifest(t, "qiantie-202611060022")})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/product-updates/latest", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Cache-Control"); got != "no-cache, must-revalidate" {
		t.Fatalf("Cache-Control = %q, want revalidation", got)
	}
	var body struct {
		ReleaseID string `json:"releaseId"`
		Version   string `json:"version"`
		Title     string `json:"title"`
		Sections  []struct {
			Type  string `json:"type"`
			Label string `json:"label"`
			Items []struct {
				Title string `json:"title"`
			} `json:"items"`
		} `json:"sections"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ReleaseID != "qiantie-202611060022" || body.Version != "17" || body.Title == "" {
		t.Fatalf("body = %#v, want the 17 release note", body)
	}
	if len(body.Sections) == 0 {
		t.Fatalf("sections = %#v, want grouped entries", body.Sections)
	}
	for _, section := range body.Sections {
		if section.Type != "feature" && section.Type != "improvement" && section.Type != "fix" {
			t.Fatalf("section type = %q, want feature/improvement/fix", section.Type)
		}
		if section.Label == "" || len(section.Items) == 0 || section.Items[0].Title == "" {
			t.Fatalf("section = %#v, want a label and items", section)
		}
	}
}

func TestProductUpdatesLatestReturnsV27RetryRecoveryNotes(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: writeReleaseManifest(t, "qiantie-202611071037")})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/product-updates/latest", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var body struct {
		ReleaseID string `json:"releaseId"`
		Version   string `json:"version"`
		Sections  []struct {
			Items []struct {
				Title string `json:"title"`
			} `json:"items"`
		} `json:"sections"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ReleaseID != "qiantie-202611071037" || body.Version != "27" {
		t.Fatalf("body = %#v, want the v27 retry recovery release note", body)
	}
	if len(body.Sections) != 1 || len(body.Sections[0].Items) != 3 {
		t.Fatalf("sections = %#v, want three retry recovery fixes", body.Sections)
	}
}

func TestProductUpdatesLatestStaysSilentForUnknownRelease(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: writeReleaseManifest(t, "qiantie-unknown")})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/product-updates/latest", nil))

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestProductUpdatesLatestStaysSilentWithoutManifest(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: filepath.Join(t.TempDir(), "missing.json")})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/product-updates/latest", nil))

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func TestProductUpdatesLatestRejectsWrites(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: writeReleaseManifest(t, "qiantie-202611060022")})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/product-updates/latest", nil))

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}
