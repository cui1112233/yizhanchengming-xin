package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestModelCatalogCRUDHidesCredential(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	create := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models", map[string]any{
		"id": "text-demo", "kind": "text", "displayName": "测试文本", "modelId": "demo", "baseUrl": "https://example.com/v1/chat/completions", "credential": "secret-value", "enabled": true,
	}, now, secret)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	if strings.Contains(create.Body.String(), "secret-value") {
		t.Fatal("model credential leaked in response")
	}
	list := doNovelFetchRequest(t, handler, http.MethodGet, "/api/models?kind=text", nil, now, secret)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Models []map[string]any `json:"models"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Models) != 1 || payload.Models[0]["hasCredential"] != true {
		t.Fatalf("unexpected public catalog: %#v", payload.Models)
	}
}
