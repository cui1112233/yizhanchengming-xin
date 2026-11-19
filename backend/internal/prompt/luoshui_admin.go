package prompt

import (
	"strings"

	"qiantie/backend/internal/prompt/luoshui"
)

// LuoshuiAdminGroup 描述后台的一个功能分区。Slot 是后台的归属 ID，
// ToolID 对应 pluginToolPrompts 里的插件工具；ToolID 为空表示未归属。
type LuoshuiAdminGroup struct {
	Slot   string
	Label  string
	ToolID string
	// IDs 用于非插件分区的固定条目列表；为空时按「未归属」规则收纳剩余条目。
	IDs []string
	// Runtime 表示该分区的正文会被漫剧链路运行时读取。
	Runtime bool
}

// 漫剧创作链路实际读取的提示词，按「剧本 → 资产 → 分镜 → 审核/概要」顺序排列。
// v88_* 三条是 V88 内置运行词（剧本医生/分镜医生/故事概要），运行时经
// luoshui.V88RuntimeBody 读取，后台发布覆盖后生效。
var comicScreenplayPromptIDs = []string{
	CreativeScript144, CreativeStoryboard307, CreativeCharacter213, CreativeScene289,
	"a1_novel_to_screenplay",
	"b_director_screenplay",
	"asset_character_v6_3d",
	"asset_character_real_v7",
	"asset_character_standard_v5",
	"asset_character_lite",
	"asset_scene_v4",
	"asset_scene_v3",
	"asset_prop_v4",
	"asset_prop_v3",
	"storyboard_xiaoluo_v7_15s",
	"storyboard_xiaoyunque_v3",
	"v88_doctor_screenplay",
	"v88_doctor_storyboard",
	"v88_synopsis",
}

// 分组顺序固定：先按 5 个插件功能，再是未归属。漫剧创作链路已拆为独立
// 后台模块「漫剧创作」（ComicScreenplayAdminModule），不在本列表里。
// 每个功能内部严格按 pluginToolPrompts 的 Go 切片顺序，保证与插件下拉一致。
var luoshuiAdminGroups = []LuoshuiAdminGroup{
	{Slot: "plugin.novel-to-drama", Label: "小说转剧", ToolID: "novel-to-drama"},
	{Slot: "plugin.storyboard-master", Label: "分镜大师", ToolID: "storyboard-master"},
	{Slot: "plugin.viral-opening", Label: "爆款开头", ToolID: "viral-opening"},
	{Slot: "plugin.character-edit", Label: "角色修改", ToolID: "character-edit"},
	{Slot: "plugin.cover-generation", Label: "封面生成", ToolID: "cover-generation"},
	{Slot: "plugin.unassigned", Label: "未归属", ToolID: ""},
}

// luoshuiBatchRuntimeIDs 是后台批量剧本链路实际读取正文的条目：
// 它们不在 5 个插件功能里，但运行时同样优先读取已发布的覆盖版本。
var luoshuiBatchRuntimeIDs = map[string]bool{
	"a1_novel_to_screenplay":      true,
	"b_director_screenplay":       true,
	"storyboard_xiaoluo_v7_15s":   true,
	"storyboard_xiaoyunque_v3":    true,
	"asset_character_v6_3d":       true,
	"asset_character_lite":        true,
	"asset_character_real_v7":     true,
	"asset_character_standard_v5": true,
	"asset_scene_v3":              true,
	"asset_scene_v4":              true,
	"asset_prop_v3":               true,
	"asset_prop_v4":               true,
}

// luoshuiBatchStepIDs 只在批次步骤表里作为 prompt_id 出现，
// 当前没有任何代码读取它们的正文，后台编辑不会改变输出。
var luoshuiBatchStepIDs = map[string]bool{
	"batch_step_0": true,
	"batch_step_1": true,
	"batch_step_2": true,
	"batch_step_3": true,
	"batch_step_4": true,
	"batch_step_5": true,
	"batch_step_6": true,
	"batch_step_7": true,
	"batch_step_8": true,
}

// LuoshuiPluginSlots 返回后台「CM 插件」模块的功能分区，复用现有
// module + slot 机制展示，不另造界面。
func LuoshuiPluginSlots() []map[string]string {
	slots := make([]map[string]string, 0, len(luoshuiAdminGroups))
	for _, group := range luoshuiAdminGroups {
		slots = append(slots, map[string]string{"id": group.Slot, "label": group.Label, "mode": "primary"})
	}
	return slots
}

// LuoshuiAdminPresets 把内嵌语料按功能转换成管理后台预设词列表。
// 5 个插件功能与未归属条目归「CM 插件」模块；漫剧创作链路条目归独立
// 「漫剧创作」模块。每个功能只放该功能实际引用的条目。
func LuoshuiAdminPresets() []map[string]any {
	registry, err := luoshui.Default()
	if err != nil {
		return nil
	}
	ids := registry.IDs()
	entries := make(map[string]luoshui.Entry, len(ids))
	for _, id := range ids {
		if entry, ok := registry.Get(id); ok {
			entries[id] = entry
		}
	}
	comicChain := make(map[string]bool, len(comicScreenplayPromptIDs))
	for _, id := range comicScreenplayPromptIDs {
		comicChain[id] = true
	}
	comicGroup := LuoshuiAdminGroup{Slot: "plugin.comic-screenplay", Label: "漫剧创作", Runtime: true}
	items := make([]map[string]any, 0, len(ids))
	assigned := make(map[string]bool, len(ids))
	index := 0
	for _, group := range luoshuiAdminGroups {
		if group.ToolID != "" {
			for _, option := range pluginToolPrompts[group.ToolID] {
				entry, ok := entries[option.corpusID]
				if !ok || assigned[option.corpusID] {
					continue
				}
				assigned[option.corpusID] = true
				items = append(items, luoshuiAdminItem(entry, group, index, false))
				index++
			}
			continue
		}
		if len(group.IDs) > 0 {
			for _, id := range group.IDs {
				entry, ok := entries[id]
				if !ok || assigned[id] {
					continue
				}
				assigned[id] = true
				items = append(items, luoshuiAdminItem(entry, group, index, false))
				index++
			}
			continue
		}
		// 未归属：漫剧创作链路条目已归属独立模块，不收进来。
		for _, id := range ids {
			if assigned[id] || comicChain[id] {
				continue
			}
			items = append(items, luoshuiAdminItem(entries[id], group, index, true))
			index++
		}
	}
	// 漫剧创作独立模块：按「剧本 → 资产 → 分镜 → 审核/概要」固定顺序。
	for _, id := range comicScreenplayPromptIDs {
		entry, ok := entries[id]
		if !ok || assigned[id] {
			continue
		}
		assigned[id] = true
		items = append(items, luoshuiAdminItem(entry, comicGroup, index, false))
		index++
	}
	return items
}

// ComicScreenplaySlots 返回「漫剧创作」独立模块的分区（单分区固定目录）。
func ComicScreenplaySlots() []map[string]string {
	return []map[string]string{{"id": "plugin.comic-screenplay", "label": "漫剧创作", "mode": "primary"}}
}

func luoshuiAdminItem(entry luoshui.Entry, group LuoshuiAdminGroup, index int, unassigned bool) map[string]any {
	// 漫剧创作链路条目归属独立后台模块；其余条目留在 CM 插件模块。
	module := LuoshuiAdminModule
	if group.Slot == "plugin.comic-screenplay" {
		module = ComicScreenplayAdminModule
	}
	return map[string]any{
		"id":                LuoshuiOverridePresetID(entry.ID),
		"module":            module,
		"name":              entry.Name,
		"kind":              "base",
		"description":       luoshuiAdminDescription(entry, group, unassigned),
		"compatibleBaseIds": []any{},
		"version":           1,
		"revision":          1,
		"status":            "published",
		"body":              entry.Body(),
		"protocolLock":      map[string]any{"slot": group.Slot},
		"createdAt":         "",
		"createdBy":         "cm-embed",
		"publishedAt":       nil,
		"publishedBy":       nil,
		"builtin":           true,
		"corpusId":          entry.ID,
		"order":             index,
	}
}

func luoshuiAdminDescription(entry luoshui.Entry, group LuoshuiAdminGroup, unassigned bool) string {
	// V88 内置运行词：正文是 V88 自研功能的内置词（非 CM v2.9.0 取证原文），
	// 运行时读取已发布覆盖版本，后台编辑发布后真实生效。
	if entry.Section == "v88" {
		parts := []string{"V88 内置运行词（非洛水取证原文）"}
		if group.Label != "" {
			parts = append(parts, group.Label+"链路在用，编辑发布后生效")
		}
		return strings.Join(parts, " · ")
	}
	parts := []string{"CM v2.9.0 内嵌原文"}
	if entry.Section != "" {
		parts = append(parts, "§"+entry.Section)
	}
	if unassigned {
		switch {
		case luoshuiBatchRuntimeIDs[entry.ID]:
			parts = append(parts, "后台批量链路在用，编辑发布后生效")
		case luoshuiBatchStepIDs[entry.ID]:
			parts = append(parts, "批次步骤表引用；当前无代码读取正文")
		default:
			parts = append(parts, "历史存档，当前无运行时引用")
		}
	} else if group.Runtime {
		parts = append(parts, group.Label+"链路在用，编辑发布后生效")
	} else if group.Label != "" {
		parts = append(parts, group.Label+"下拉选项")
	}
	if entry.HiddenFromUser() {
		parts = append(parts, "规格要求不向普通用户展示正文")
	}
	return strings.Join(parts, " · ")
}

// LuoshuiAdminPreset 按 ID 和版本返回内嵌基线；只存在 V1。
func LuoshuiAdminPreset(id string, version int) (map[string]any, bool) {
	if version != 1 {
		return nil, false
	}
	for _, item := range LuoshuiAdminPresets() {
		if stringValue(item["id"]) == id {
			return item, true
		}
	}
	return nil, false
}

// MergeLuoshuiAdminPresets 把内嵌基线与管理员覆盖版本合并成「CM 插件」视图：
// 只含 CM 插件模块的内嵌条目，漫剧创作条目归独立模块视图。
// 同一 ID 的覆盖版本优先；删除覆盖记录后自然回退到内嵌原文。
func MergeLuoshuiAdminPresets(overrides []map[string]any) []map[string]any {
	return mergeLuoshuiAdminByModule(LuoshuiAdminModule, overrides, true)
}

// MergeComicScreenplayAdminPresets 把内嵌基线与管理员覆盖版本合并成
// 「漫剧创作」视图。固定目录不追加自建条目；覆盖记录按预设词 ID 匹配，
// 不限历史模块归属（旧「CM 插件」模块下的覆盖记录同样生效）。
func MergeComicScreenplayAdminPresets(overrides []map[string]any) []map[string]any {
	return mergeLuoshuiAdminByModule(ComicScreenplayAdminModule, overrides, false)
}

// mergeLuoshuiAdminByModule 按 module 圈定内嵌基线后合并覆盖版本。
// appendExtras=true 时（CM 插件视图）把未匹配基线的覆盖记录追加在尾部，
// 兼容历史自建条目；false（漫剧创作视图）则丢弃非基线条目。
func mergeLuoshuiAdminByModule(module string, overrides []map[string]any, appendExtras bool) []map[string]any {
	builtinAll := LuoshuiAdminPresets()
	builtin := make([]map[string]any, 0, len(builtinAll))
	for _, item := range builtinAll {
		if stringValue(item["module"]) == module {
			builtin = append(builtin, item)
		}
	}
	if len(builtin) == 0 {
		return overrides
	}
	builtinByID := make(map[string]map[string]any, len(builtin))
	order := make([]string, 0, len(builtin))
	for _, item := range builtin {
		id := stringValue(item["id"])
		builtinByID[id] = item
		order = append(order, id)
	}
	overridesByID := make(map[string][]map[string]any)
	extra := make([]map[string]any, 0)
	for _, item := range overrides {
		id := stringValue(item["id"])
		if _, ok := builtinByID[id]; ok {
			overridesByID[id] = append(overridesByID[id], item)
			continue
		}
		extra = append(extra, item)
	}
	result := make([]map[string]any, 0, len(builtin)+len(overrides))
	for _, id := range order {
		base := builtinByID[id]
		replaced := false
		for _, item := range overridesByID[id] {
			if intValue(item["version"]) >= intValue(base["version"]) {
				result = append(result, lockPresetSlot(item, base))
				replaced = true
			}
		}
		if !replaced {
			result = append(result, base)
		}
	}
	if appendExtras {
		result = append(result, extra...)
	}
	// Existing published records may still have the previous display names.
	// Keep IDs and prompt bodies intact so saved projects continue to resolve.
	renames := strings.NewReplacer("小洛", "小Y", "小罗", "小Y", "洛水", "CM")
	for i, item := range result {
		copy := make(map[string]any, len(item))
		for key, value := range item {
			copy[key] = value
		}
		for _, field := range []string{"name", "description"} {
			if text, ok := copy[field].(string); ok {
				copy[field] = renames.Replace(text)
			}
		}
		result[i] = copy
	}
	return result
}

// lockPresetSlot 保证已发布覆盖版本仍显示在代码固定的分区里；
// 否则旧分区名会跟着覆盖记录一起被带到后台列表。
func lockPresetSlot(item, base map[string]any) map[string]any {
	baseLock, ok := base["protocolLock"].(map[string]any)
	if !ok {
		return item
	}
	slot := stringValue(baseLock["slot"])
	if slot == "" {
		return item
	}
	merged := make(map[string]any, len(item))
	for key, value := range item {
		merged[key] = value
	}
	lock := make(map[string]any)
	if current, ok := item["protocolLock"].(map[string]any); ok {
		for key, value := range current {
			lock[key] = value
		}
	}
	lock["slot"] = slot
	merged["protocolLock"] = lock
	return merged
}
