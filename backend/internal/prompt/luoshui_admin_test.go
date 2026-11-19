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
	// 漫剧创作已拆为独立后台模块，CM 插件只剩 5 个功能 + 未归属。
	if slots := LuoshuiPluginSlots(); len(slots) != 6 {
		t.Fatalf("slots=%d want=6", len(slots))
	}
	if slots := ComicScreenplaySlots(); len(slots) != 1 || slots[0]["id"] != "plugin.comic-screenplay" {
		t.Fatalf("comic slots=%v", slots)
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
		"plugin.comic-screenplay":  19,
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
	// 模块归属：漫剧创作条目在独立模块，CM 插件条目留在 CM 插件模块。
	for _, item := range LuoshuiAdminPresets() {
		lock, _ := item["protocolLock"].(map[string]any)
		slot, _ := lock["slot"].(string)
		module, _ := item["module"].(string)
		if slot == "plugin.comic-screenplay" && module != ComicScreenplayAdminModule {
			t.Fatalf("漫剧创作条目模块错误: %+v", item)
		}
		if slot != "plugin.comic-screenplay" && module != LuoshuiAdminModule {
			t.Fatalf("CM 插件条目模块错误: %+v", item)
		}
	}
	for _, item := range LuoshuiAdminPresets() {
		name, _ := item["name"].(string)
		if strings.Contains(name, "洛水") || strings.Contains(name, "小洛") {
			t.Fatalf("后台名称不应出现旧产品名: %q", name)
		}
	}
}

// 模块拆分后的视图合并：漫剧创作视图吃独立模块基线 + 历史 CM 插件模块下的覆盖；
// CM 插件视图不再混入漫剧创作条目。
func TestMergePresetsSplitByModule(t *testing.T) {
	all := LuoshuiAdminPresets()
	comicCount := 0
	for _, item := range all {
		if item["module"] == ComicScreenplayAdminModule {
			comicCount++
		}
	}
	if comicCount != 19 {
		t.Fatalf("comic builtin=%d want=19", comicCount)
	}
	legacy := map[string]any{
		"id":      LuoshuiOverridePresetID("storyboard_xiaoluo_v7_15s"),
		"module":  LuoshuiAdminModule,
		"name":    "15秒分镜",
		"version": 1,
		"status":  "published",
		"body":    "历史覆盖正文",
	}
	comic := MergeComicScreenplayAdminPresets([]map[string]any{legacy})
	found := false
	for _, item := range comic {
		if item["id"] == "luoshui.storyboard_xiaoluo_v7_15s" {
			found = true
			if item["body"] != "历史覆盖正文" {
				t.Fatalf("历史覆盖未生效: %+v", item)
			}
		}
	}
	if !found {
		t.Fatal("漫剧创作视图缺少分镜词条目")
	}
	plugin := MergeLuoshuiAdminPresets(nil)
	for _, item := range plugin {
		if item["module"] == ComicScreenplayAdminModule {
			t.Fatalf("CM 插件视图不应出现漫剧创作条目: %+v", item)
		}
	}
	// 非基线条目的覆盖在漫剧创作固定目录中被丢弃，不追加。
	stray := MergeComicScreenplayAdminPresets([]map[string]any{{
		"id": "luoshui.plugin_cover_design", "module": LuoshuiAdminModule, "version": 1, "status": "published", "body": "x",
	}})
	if len(stray) != comicCount {
		t.Fatalf("stray len=%d want=%d", len(stray), comicCount)
	}
}

func TestMergeLuoshuiAdminPresetsPrefersOverrides(t *testing.T) {
	items := LuoshuiAdminPresets()
	// CM 插件视图只含 CM 插件模块的基线条目；漫剧创作条目已拆独立模块。
	pluginCount := 0
	for _, item := range items {
		if item["module"] == LuoshuiAdminModule {
			pluginCount++
		}
	}
	override := map[string]any{
		"id":          LuoshuiOverridePresetID("plugin_cover_design"),
		"module":      LuoshuiAdminModule,
		"name":        "洛水小洛封面设计",
		"description": "洛水小罗工具",
		"version":     1,
		"revision":    2,
		"status":      "draft",
		"body":        "管理员草稿",
	}
	merged := MergeLuoshuiAdminPresets([]map[string]any{override})
	// 覆盖命中基线条目时原位替换，列表长度不变；只有未匹配基线的
	// 自建条目才会追加在尾部（本用例没有）。
	if len(merged) != pluginCount {
		t.Fatalf("len=%d want=%d", len(merged), pluginCount)
	}
	found := false
	for _, item := range merged {
		if item["id"] == "luoshui.plugin_cover_design" {
			found = true
			if item["name"] != "CM小Y封面设计" || item["description"] != "CM小Y工具" {
				t.Fatalf("旧名称未兼容: %+v", item)
			}
			if override["name"] != "洛水小洛封面设计" {
				t.Fatal("合并不应修改输入记录")
			}
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
		"module":       ComicScreenplayAdminModule,
		"name":         "小说转漫剧 A1",
		"version":      1,
		"revision":     2,
		"status":       "published",
		"body":         "管理员改过的正文",
		"protocolLock": map[string]any{"slot": "plugin.unassigned", "format": "text"},
	}
	merged := MergeComicScreenplayAdminPresets([]map[string]any{override})
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
