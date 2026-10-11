package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestNovelFetchWorkshopConfigHidesCredentialsAndPreservesOnEdit(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"api_key": "secret-value", "model": "old-model"}},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "secret-value") {
		t.Fatalf("config leaked credential: status=%d body=%s", get.Code, get.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	app := payload["appConfig"].(map[string]any)
	ai := app["ai"].(map[string]any)
	if ai["api_key"] != nil || ai["api_key_configured"] != true {
		t.Fatalf("unexpected public ai config: %#v", ai)
	}
	put := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"model": "new-model", "api_key": ""}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	saved, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	savedAI := saved["appConfig"].(map[string]any)["ai"].(map[string]any)
	if savedAI["api_key"] != "secret-value" || savedAI["model"] != "new-model" {
		t.Fatalf("edit should preserve credential: %#v", savedAI)
	}
	put = doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"model": "put-model", "api_key": ""}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
}
