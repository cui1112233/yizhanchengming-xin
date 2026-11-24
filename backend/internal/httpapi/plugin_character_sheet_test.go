package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
	"time"
)

func TestCharacterSheetPromptIsBuiltOnlyFromPublishedOwnedTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{"id":"sheet","module":"shuihuo-production","kind":"base","status":"published","body":"后端三视图布局","protocolLock":{"slot":"shuihuo.asset.character-sheet"}},{"id":"private","module":"other","kind":"base","status":"published","body":"不可使用"},{"id":"draft","module":"shuihuo-production","kind":"base","status":"draft","body":"不可使用","protocolLock":{"slot":"shuihuo.asset.character-sheet"}}]`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, PresetCatalog: prompt.NewCatalog(path)})
	for _, tc := range []struct {
		id   string
		code int
	}{{"sheet", 200}, {"private", 400}, {"draft", 400}, {"missing", 400}} {
		rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/plugins/character-turnaround/prompt", map[string]any{"presetId": tc.id, "description": "浅蓝外套青年"})
		if rec.Code != tc.code {
			t.Fatalf("%s: status=%d body=%s", tc.id, rec.Code, rec.Body.String())
		}
		if tc.code == 200 && (!strings.Contains(rec.Body.String(), "后端三视图布局") || !strings.Contains(rec.Body.String(), "浅蓝外套青年")) {
			t.Fatal(rec.Body.String())
		}
	}
}
