package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt/luoshui"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestStandaloneAssetReparseRouteUsesSavedModel(t *testing.T) {
	now := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	_, _ = store.CreateProject(context.Background(), screenplay.Project{ID: "p", OwnerUsername: "alice", ModelID: "our-configured-model", PermissionMode: screenplay.PermissionBypass, StyleDesc: "do-not-inject"})
	calls := 0
	provider := screenplayProviderFunc(func(ctx context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
		calls++
		system, user, _ := luoshui.NativeAssetMessages("character", "完整原文")
		if r.SystemPrompt != system || r.UserPrompt != user || r.Temperature != 1 || r.MaxTokens != 20000 {
			t.Fatal("wrong final provider messages or parameters")
		}
		return `{"characters":[{"name":"甲","prompt":"新提示词"}]}`, nil
	})
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || model != "our-configured-model" {
			t.Fatalf("wrong model source: %s %s", owner, model)
		}
		return provider, nil
	}, func() time.Time { return now })
	for _, mode := range []string{"standalone", "unknown"} {
		r := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/p/assets/extract", map[string]any{"mode": mode, "type": "character", "content": "完整原文", "model": "ignored-model", "confirmed": true})
		if mode == "standalone" && r.Code != 200 {
			t.Fatalf("%d %s", r.Code, r.Body.String())
		}
		if mode == "unknown" && r.Code < 400 {
			t.Fatalf("unknown entrance accepted: %d", r.Code)
		}
	}
	if calls != 1 {
		t.Fatalf("provider calls=%d", calls)
	}
	rows, _ := store.ListAssets(context.Background(), "alice", "p", screenplay.AssetCharacter)
	if len(rows) != 1 || rows[0].Name != "甲" {
		t.Fatalf("not persisted: %+v", rows)
	}
}
