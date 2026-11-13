package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
	"testing"
	"time"
)

func TestWorkshopImportedKnowledgeRepairIsScopedVersionedAndPrivate(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	settings, _ := store.GetConfig(context.Background(), "tester")
	high := settings["knowledgeLibraries"].(map[string]any)["high_imitation"].(map[string]any)
	depths := high["settings"].(map[string]any)["depths"].([]any)
	for i := range depths {
		depths[i] = "??"
	}
	if err := store.PutConfig(context.Background(), "tester", settings); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1800000000, 0)
	h := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchStore: store})
	path := "/api/novel-fetch-workshop/knowledge/repair-imported"
	check := doNovelFetchRequest(t, h, "POST", path, map[string]any{"apply": false}, now, "secret")
	if check.Code != 200 {
		t.Fatalf("dry run: %d %s", check.Code, check.Body)
	}
	var result map[string]any
	if err := json.Unmarshal(check.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["report"].(map[string]any)["repaired"].(float64) < 3 {
		t.Fatal("repair not detected")
	}
	current, _ := store.GetConfig(context.Background(), "tester")
	if current["importedKnowledgeRepairBackup"] != nil {
		t.Fatal("dry run mutated")
	}
	stale := doNovelFetchRequest(t, h, "POST", path, map[string]any{"apply": true, "configRevision": "stale"}, now, "secret")
	if stale.Code != http.StatusConflict {
		t.Fatal("stale repair accepted")
	}
	applied := doNovelFetchRequest(t, h, "POST", path, map[string]any{"apply": true, "configRevision": result["configRevision"]}, now, "secret")
	if applied.Code != 200 {
		t.Fatalf("apply: %d %s", applied.Code, applied.Body)
	}
	current, _ = store.GetConfig(context.Background(), "tester")
	if current["importedKnowledgeRepairBackup"] == nil {
		t.Fatal("missing recovery backup")
	}
	public, _ := json.Marshal(publicNovelFetchConfig(current))
	if strings.Contains(string(public), "importedKnowledgeRepairBackup") {
		t.Fatal("private backup exposed")
	}
	other, _ := store.GetConfig(context.Background(), "other")
	if other["importedKnowledgeRepairBackup"] != nil {
		t.Fatal("other owner modified")
	}
}
