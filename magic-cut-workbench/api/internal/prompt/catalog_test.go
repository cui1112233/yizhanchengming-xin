package prompt

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCatalogDistinguishesObservedTextFromBehavioralReplacements(t *testing.T) {
	observed, ok := Catalog{}.Get("magic_cut.ai_pre_roll.negative_constraints.v1")
	if !ok {
		t.Fatal("observed prompt is missing")
	}
	if observed.SourceAuthority != "observed-client" || observed.SourceLocation == "" {
		t.Fatalf("observed = %+v", observed)
	}

	derived, ok := Catalog{}.Get("magic_cut.hot_replication.infer_video.v1")
	if !ok {
		t.Fatal("derived prompt is missing")
	}
	if derived.SourceAuthority != "derived-equivalent" || len(derived.RequiredVariables) == 0 {
		t.Fatalf("derived = %+v", derived)
	}
}

func TestRegisteredSystemPromptsNeverAppearInFrontendSource(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	frontendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "前台", "src"))
	fingerprints := Catalog{}.SystemFingerprints()

	err := filepath.WalkDir(frontendRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, fingerprint := range fingerprints {
			if strings.Contains(string(data), fingerprint) {
				t.Fatalf("frontend file %s contains a backend system prompt", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
