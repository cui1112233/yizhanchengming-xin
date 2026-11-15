package prompt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestRecoveredCatalogPinsAll63ExactBodiesAndAliases(t *testing.T) {
	manifest, err := recoveredCreativeFiles.ReadFile("screenplay_recovered_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	// Pins the reviewed per-record digests and protocol decisions independently
	// of the loader: accidental edits to both a body and its digest still fail.
	if got := fmt.Sprintf("%x", sha256.Sum256(manifest)); got != "aec52f679d50167b8bdd8a643d61967ac52716b61be1ab4b46c0a3d8954581e4" {
		t.Fatal("reviewed catalog changed", got)
	}
	items, err := RecoveredCreativePromptCatalog()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, m := range items {
		counts[m.Type]++
		for _, id := range []string{m.ID, strconv.Itoa(m.SourceID)} {
			p, err := ResolveCreativePrompt(id, m.Type)
			if err != nil {
				t.Fatal(id, err)
			}
			if p.ID != m.ID || p.SourceID != m.SourceID || p.Purpose != m.Type || p.Version != m.Digest[:16] || p.NormalizedDigest != m.NormalizedDigest {
				t.Fatal("snapshot identity changed", id)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(p.Body))) != m.Digest || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(p.Body, "\r\n", "\n")))) != m.NormalizedDigest {
				t.Fatal("body changed", id)
			}
			if _, err := ResolveCreativePrompt(id, "wrong-purpose"); err == nil {
				t.Fatal("wrong purpose accepted", id)
			}
		}
		if m.OutputProtocol == "" || m.ParserVersion == "" || m.InputSource == "" || len(m.ModelCapabilities) == 0 || m.Source == "" {
			t.Fatal("missing contract", m.ID)
		}
		if strings.Contains(m.Name, "小洛") || strings.Contains(m.Name, "洛水") {
			t.Fatal("old visible name", m.Name)
		}
		if !m.Supported && m.UnsupportedReason == "" {
			t.Fatal("missing unsupported reason", m.ID)
		}
	}
	for kind, want := range map[string]int{"script": 3, "storyboard": 41, "character-design": 12, "scene-design": 6, "voice-design": 1} {
		if counts[kind] != want {
			t.Fatal(kind, counts[kind], want)
		}
	}
	for _, id := range []string{"", "0", "999999", "deleted_prompt"} {
		if _, err := ResolveCreativePrompt(id, "storyboard"); err == nil {
			t.Fatal("unknown accepted", id)
		}
	}
}

func TestRecoveredProtocolsKeepDistinctFieldsReportsAndDurations(t *testing.T) {
	for _, tc := range []struct {
		id, protocol, mode string
		min, max, fixed    float64
		supported          bool
	}{
		{"137", "storyboard_image_voice_v1", "json_with_report", 0, 0, 0, true},
		{"138", "storyboard_image_voice_v1", "json_with_report", 0, 0, 0, false},
		{"139", "storyboard_image_voice_v1", "json_object", 15, 15, 15, true},
		{"147", "storyboard_seedance_timed_v1", "json_object", 2, 15, 0, true},
		{"175", "storyboard_nine_grid_v1", "json_object", 15, 15, 15, true},
		{"190", "storyboard_seedance_timed_v1", "json_object", 10, 10, 10, true},
		{"240", "storyboard_video_props_v1", "json_object", 0, 14.5, 0, true},
		{"268", "storyboard_video_props_v1", "json_object", 10, 15, 0, true},
		{"303", "storyboard_video_v1", "json_object", 15, 15, 15, true},
		{"307", "storyboard_video_props_v1", "json_object", 4, 30, 0, true},
		{"310", "storyboard_video_props_v1", "json_object", 0, 15, 0, true},
	} {
		p, err := ResolveCreativePrompt(tc.id, "storyboard")
		if err != nil {
			t.Fatal(err)
		}
		if p.OutputProtocol != tc.protocol || p.ResponseMode != tc.mode || p.MinDurationSec != tc.min || p.MaxDurationSec != tc.max || p.FixedDurationSec != tc.fixed || p.Supported != tc.supported {
			t.Fatalf("%s contract=%+v", tc.id, p.CreativePromptContract)
		}
	}
	for _, id := range []string{"144", "206", "293"} {
		p, err := ResolveCreativePrompt(id, "script")
		if err != nil {
			t.Fatal(err)
		}
		if p.ResponseMode != "json_with_report" || len(p.ReportMarkers) == 0 {
			t.Fatal("script report lost", id)
		}
		if id == "293" && (p.Supported || len(p.ReportMarkers) != 2) {
			t.Fatal("batch script incorrectly enabled")
		}
	}
	nine, _ := ResolveCreativePrompt("175", "storyboard")
	for _, field := range []string{"image_desc", "image_prompt", "content_density", "timecode_global", "source_elements", "voice_actor", "emotion", "line"} {
		if !strings.Contains(strings.Join(nine.RequiredFields, ","), field) {
			t.Fatal("nine-grid field lost", field)
		}
	}
}

func TestRecoveredCatalogCannotBeMutatedThroughMetadata(t *testing.T) {
	items, err := RecoveredCreativePromptCatalog()
	if err != nil {
		t.Fatal(err)
	}
	items[1].RequiredFields[0] = "corrupted"
	p, err := ResolveCreativePrompt("137", "storyboard")
	if err != nil {
		t.Fatal(err)
	}
	if p.RequiredFields[0] == "corrupted" {
		t.Fatal("shared mutable catalog")
	}
	b, err := json.Marshal(CreativePromptOptions())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "\"body\"") {
		t.Fatal("catalog leaked body")
	}
	seen := map[string]bool{}
	for _, m := range CreativePromptOptions() {
		if seen[m.ID] {
			t.Fatal("duplicate", m.ID)
		}
		seen[m.ID] = true
		if m.SourceID == 127 {
			t.Fatal("voice belongs to separate service")
		}
	}
}
