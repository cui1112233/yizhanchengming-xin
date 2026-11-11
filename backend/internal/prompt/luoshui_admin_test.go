package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/prompt/luoshui"
)

func TestLuoshuiAdminPresetsExposeEmbeddedCorpus(t *testing.T) {
	items := LuoshuiAdminPresets()
	if len(items) == 0 {
		t.Fatal("内嵌洛水语料未加载")
	}
	var cover map[string]any
	for _, item := range items {
		if item["id"] == "luoshui.plugin_cover_design" {
			cover = item
			break
		}
	}
	if cover == nil {
		t.Fatal("缺少封面设计条目")
	}
	if cover["module"] != LuoshuiAdminModule || cover["status"] != "published" || cover["builtin"] != true {
		t.Fatalf("cover=%+v", cover)
	}
	if body, _ := cover["body"].(string); !strings.Contains(body, "番茄小说、七猫免费小说平台顶级封面设计专家") {
		t.Fatalf("正文未包含洛水原文: %q", body)
	}
}

func TestLuoshuiAdminPresetsGroupedByFunction(t *testing.T) {
	if slots := LuoshuiPluginSlots(); len(slots) != 7 {
		t.Fatalf("slots=%d want=7", len(slots))
	}
	counts := make(map[string]int)
	orders := make(map[string][]string)
	for _, item := range LuoshuiAdminPresets() {
		lock, _ := item["protocolLock"].(map[string]any)
		slot, _ := lock["slot"].(string)
		counts[slot]++
		orders[slot] = append(orders[slot], item["corpusId"].(string))
	}
	wantCounts := map[string]int{
		"plugin.novel-to-drama":    8,
		"plugin.storyboard-master": 2,
		"plugin.viral-opening":     2,
		"plugin.character-edit":    1,
		"plugin.cover-generation":  1,
		"plugin.comic-screenplay":  12,
		"plugin.unassigned":        23,
	}
	for slot, want := range wantCounts {
		if counts[slot] != want {
			t.Fatalf("slot=%s got=%d want=%d", slot, counts[slot], want)
		}
	}
	wantOrder := make([]string, 0)
	for _, option := range pluginToolPrompts["novel-to-drama"] {
		wantOrder = append(wantOrder, option.corpusID)
	}
	if strings.Join(orders["plugin.novel-to-drama"], ",") != strings.Join(wantOrder, ",") {
		t.Fatalf("小说转剧顺序错误: got=%v want=%v", orders["plugin.novel-to-drama"], wantOrder)
	}
	if strings.Join(orders["plugin.comic-screenplay"], ",") != strings.Join(comicScreenplayPromptIDs, ",") {
		t.Fatalf("漫剧创作顺序错误: got=%v want=%v", orders["plugin.comic-screenplay"], comicScreenplayPromptIDs)
	}
	for _, item := range LuoshuiAdminPresets() {
		name, _ := item["name"].(string)
		if strings.Contains(name, "洛水") || strings.Contains(name, "小洛") {
			t.Fatalf("后台名称不应出现旧产品名: %q", name)
		}
	}
}

func TestMergeLuoshuiAdminPresetsPrefersOverrides(t *testing.T) {
	items := LuoshuiAdminPresets()
	override := map[string]any{
		"id":       LuoshuiOverridePresetID("plugin_cover_design"),
		"module":   LuoshuiAdminModule,
		"name":     "封面设计",
		"version":  1,
		"revision": 2,
		"status":   "draft",
		"body":     "管理员草稿",
	}
	merged := MergeLuoshuiAdminPresets([]map[string]any{override})
	if len(merged) != len(items) {
		t.Fatalf("len=%d want=%d", len(merged), len(items))
	}
	found := false
	for _, item := range merged {
		if item["id"] == "luoshui.plugin_cover_design" {
			found = true
			if item["body"] != "管理员草稿" || item["status"] != "draft" {
				t.Fatalf("覆盖未生效: %+v", item)
			}
		}
	}
	if !found {
		t.Fatal("合并后缺少封面设计条目")
	}
}

func TestMergeLuoshuiAdminPresetsLocksComicScreenplaySlot(t *testing.T) {
	override := map[string]any{
		"id":           LuoshuiOverridePresetID("a1_novel_to_screenplay"),
		"module":       LuoshuiAdminModule,
		"name":         "小说转漫剧 A1",
		"version":      1,
		"revision":     2,
		"status":       "published",
		"body":         "管理员改过的正文",
		"protocolLock": map[string]any{"slot": "plugin.unassigned", "format": "text"},
	}
	merged := MergeLuoshuiAdminPresets([]map[string]any{override})
	for _, item := range merged {
		if item["id"] != "luoshui.a1_novel_to_screenplay" {
			continue
		}
		lock, _ := item["protocolLock"].(map[string]any)
		if lock["slot"] != "plugin.comic-screenplay" {
			t.Fatalf("slot=%v want=plugin.comic-screenplay", lock["slot"])
		}
		if lock["format"] != "text" {
			t.Fatalf("format=%v want=text", lock["format"])
		}
		return
	}
	t.Fatal("合并后缺少 a1_novel_to_screenplay")
}

func TestLuoshuiOverrideFeedsPluginPrompt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	raw := `[{"id":"luoshui.plugin_cover_design","module":"luoshui-plugin","name":"封面设计","version":1,"revision":2,"status":"published","body":"覆盖后的封面提示词"}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	SetLuoshuiOverrideCatalog(NewCatalog(path))
	t.Cleanup(func() { SetLuoshuiOverrideCatalog(nil) })

	body, ok := PluginToolSystemPrompt("cover-generation", "cover-design")
	if !ok || body != "覆盖后的封面提示词" {
		t.Fatalf("body=%q ok=%v", body, ok)
	}
	// 未被覆盖的工具仍走内嵌洛水原文。
	if storyboard, ok := PluginToolSystemPrompt("storyboard-master", "default"); !ok || strings.TrimSpace(storyboard) == "" {
		t.Fatal("未覆盖工具应继续返回内嵌原文")
	}
}

func TestLuoshuiOverrideFeedsBatchRuntimePrompt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	raw := `[{"id":"luoshui.batch_step_5","module":"luoshui-plugin","name":"批次 step 5","version":1,"revision":2,"status":"published","body":"覆盖后的分镜步骤"}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	SetLuoshuiOverrideCatalog(NewCatalog(path))
	t.Cleanup(func() { SetLuoshuiOverrideCatalog(nil) })

	body, ok := luoshui.MustDefault().RuntimeBody("batch_step_5")
	if !ok || body != "覆盖后的分镜步骤" {
		t.Fatalf("批量链路未读到覆盖: body=%q ok=%v", body, ok)
	}
	// 未被覆盖的批量条目继续回退内嵌原文。
	fallback, ok := luoshui.MustDefault().RuntimeBody("asset_scene_v3")
	if !ok || strings.TrimSpace(fallback) == "" {
		t.Fatal("未覆盖条目应回退内嵌原文")
	}
}
