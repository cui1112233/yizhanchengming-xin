package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
	"testing"
	"time"
)

func TestNovelFetchPlatformCatalogRestoresDamagedNamesWithoutOverwritingCustomLabels(t *testing.T) {
	cfg := map[string]any{"platforms": []any{map[string]any{"id": 2, "name": "????"}, map[string]any{"id": "3", "name": "七猫自定义"}, map[string]any{"id": "4", "name": "????", "visible": false}, map[string]any{"id": "999", "name": "????"}}}
	rows := novelFetchPlatformCatalog(cfg)
	if len(rows) != 3 || rows[0].(map[string]any)["name"] != "番茄付费" || rows[1].(map[string]any)["name"] != "七猫自定义" || rows[2].(map[string]any)["visible"] != false {
		t.Fatalf("catalog=%#v", rows)
	}
	if cfg["platforms"].([]any)[0].(map[string]any)["name"] != "????" {
		t.Fatal("read normalization mutated stored object")
	}
}

func TestNovelFetchPlatformCatalogRestoresEscapedNamesWithoutBackslashes(t *testing.T) {
	cfg := map[string]any{"platforms": []any{
		map[string]any{"id": "1", "name": "u9ed1u5ca9u4ed8u8d39"},
		map[string]any{"id": "999", "name": "\\u756au8304u4ed8u8d39"},
		map[string]any{"id": "42", "name": "MyBookStore"},
	}}
	rows := novelFetchPlatformCatalog(cfg)
	if len(rows) != 3 {
		t.Fatalf("catalog=%#v", rows)
	}
	if got := rows[0].(map[string]any)["name"]; got != "黑岩付费" {
		t.Fatalf("backslash-stripped name not restored: %#v", got)
	}
	if got := rows[1].(map[string]any)["name"]; got != "番茄付费" {
		t.Fatalf("escaped name not restored: %#v", got)
	}
	if got := rows[2].(map[string]any)["name"]; got != "MyBookStore" {
		t.Fatalf("normal name mutated: %#v", got)
	}
	if cfg["platforms"].([]any)[0].(map[string]any)["name"] != "u9ed1u5ca9u4ed8u8d39" {
		t.Fatal("read normalization mutated stored object")
	}
}

func TestNovelFetchBothCatalogEndpointsRepairLegacyQuestionMarks(t *testing.T) {
	now := time.Unix(1800000000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(context.Background(), "tester", map[string]any{"platforms": []any{map[string]any{"id": "2", "name": "????"}, map[string]any{"id": "1", "name": "u9ed1u5ca9u4ed8u8d39"}}})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	for _, path := range []string{"/api/novel-fetch-workshop/platforms", "/api/novel-fetch-workshop/config"} {
		r := doNovelFetchRequest(t, handler, http.MethodGet, path, nil, now, secret)
		if r.Code != 200 || !strings.Contains(r.Body.String(), "番茄付费") || !strings.Contains(r.Body.String(), "黑岩付费") ||
			strings.Contains(r.Body.String(), `"name":"????"`) || strings.Contains(r.Body.String(), "u9ed1u5ca9") {
			t.Fatalf("%s returned damaged names: %d %s", path, r.Code, r.Body.String())
		}
	}
}
