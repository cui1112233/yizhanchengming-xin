package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/prompt"
)

// 「漫剧创作」独立后台模块：slots 接口返回漫剧创作分区；presets 接口跨模块
// 合并（历史 CM 插件模块下的覆盖记录照样生效），且不再混入 CM 插件条目。
func TestComicScreenplayAdminModuleRoutes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	raw := `[{"id":"luoshui.v88_doctor_screenplay","module":"luoshui-plugin","name":"剧本医生系统词","version":1,"revision":2,"status":"published","body":"历史模块下的医生覆盖"}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	catalog := prompt.NewCatalog(path)
	now := time.Date(2026, 11, 19, 8, 0, 0, 0, time.UTC)
	store := &governanceAuthStore{accounts: map[string]auth.Account{
		"owner": {Username: "owner", Active: true, IsOwner: true},
	}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AuthStore: store, MemberProfileStore: NewMemberProfileStore(t.TempDir()), PresetCatalog: catalog})

	// slots：漫剧创作单分区。
	slots := signedJSONRequest(t, api, now, "owner", http.MethodGet, "/api/admin/preset-slots?module=comic-screenplay", nil)
	if slots.Code != http.StatusOK || !strings.Contains(slots.Body.String(), "plugin.comic-screenplay") {
		t.Fatalf("slots=%d %s", slots.Code, slots.Body.String())
	}
	cmSlots := signedJSONRequest(t, api, now, "owner", http.MethodGet, "/api/admin/preset-slots?module=luoshui-plugin", nil)
	if cmSlots.Code != http.StatusOK || strings.Contains(cmSlots.Body.String(), "plugin.comic-screenplay") {
		t.Fatalf("CM 插件 slots 不应再含漫剧创作分区: %s", cmSlots.Body.String())
	}

	// presets：漫剧创作视图跨模块合并历史覆盖；不含 CM 插件条目。
	presets := signedJSONRequest(t, api, now, "owner", http.MethodGet, "/api/admin/presets?module=comic-screenplay", nil)
	if presets.Code != http.StatusOK {
		t.Fatalf("presets=%d %s", presets.Code, presets.Body.String())
	}
	body := presets.Body.String()
	if !strings.Contains(body, "历史模块下的医生覆盖") {
		t.Fatal("历史覆盖未在漫剧创作视图生效: " + body)
	}
	if strings.Contains(body, "luoshui.plugin_cover_design") || strings.Contains(body, "封面设计") {
		t.Fatal("漫剧创作视图不应包含 CM 插件条目")
	}
	if !strings.Contains(body, "luoshui.v88_synopsis") || !strings.Contains(body, "V88 内置运行词") {
		t.Fatal("漫剧创作视图缺少 V88 内置运行词条目")
	}
}
