package prompt

import (
	"fmt"
	"strings"
	"sync/atomic"

	"qiantie/backend/internal/prompt/luoshui"
)

type PluginPromptOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	body     string
	corpusID string
}

// LuoshuiAdminModule 是管理后台「CM 插件」分区的模块名。内嵌语料作为只读
// 基线，管理员保存的版本以同一 corpusID 为键叠加，插件引用顺序不受影响。
const LuoshuiAdminModule = "luoshui-plugin"

var luoshuiOverrideCatalog atomic.Pointer[Catalog]

// SetLuoshuiOverrideCatalog 在路由装配时注入管理后台使用的预设词目录；
// 未注入时（例如单元测试）插件继续使用内嵌原文。目录同时注册给 luoshui
// 运行时，让批量剧本链路也能读到已发布的覆盖正文。
func SetLuoshuiOverrideCatalog(catalog *Catalog) {
	luoshuiOverrideCatalog.Store(catalog)
	luoshui.SetOverrideResolver(luoshuiOverrideBody)
}

// LuoshuiOverridePresetID 把洛水语料 ID 映射为管理后台可编辑的预设词 ID。
func LuoshuiOverridePresetID(corpusID string) string {
	return "luoshui." + strings.TrimSpace(corpusID)
}

// luoshuiOverrideBody 读取已发布的管理员覆盖版本。只有发布状态且正文非空的
// 记录才生效，草稿和归档记录仍回退到内嵌原文。
func luoshuiOverrideBody(corpusID string) (string, bool) {
	catalog := luoshuiOverrideCatalog.Load()
	if catalog == nil {
		return "", false
	}
	record, found, err := catalog.PublishedByID(LuoshuiOverridePresetID(corpusID))
	if err != nil || !found || record.Module != LuoshuiAdminModule {
		return "", false
	}
	body := strings.TrimSpace(record.Body)
	return body, body != ""
}

// 4 条文字类工具的提示词选项顺序与洛水 v2.9.0 插件面板一致；
// 正文直接引用 internal/prompt/luoshui 语料里的洛水原文，不在代码里重写或编造。
var pluginToolPrompts = map[string][]PluginPromptOption{
	"novel-to-drama": {
		{ID: "minimax-emotion-2", Name: "【Minimax情绪专用2】-不会修改原文", corpusID: "plugin_minimax_emotion2_no_rewrite"},
		{ID: "minimax-emotion-short-novel", Name: "【Minimax情绪专用】-短篇长篇小说转剧本", corpusID: "plugin_minimax_emotion_short_novel"},
		{ID: "novel-no-rewrite", Name: "小说转剧-不会修改原文", corpusID: "plugin_novel_to_script_no_rewrite"},
		{ID: "mid-length", Name: "中长篇", corpusID: "plugin_mid_length_script"},
		{ID: "emotion-short-novel", Name: "【带情绪】短篇长篇小说转剧本", corpusID: "plugin_emotion_short_novel_to_script"},
		{ID: "mid-length-rewrite", Name: "中篇改剧本提示词", corpusID: "plugin_mid_length_rewrite"},
		{ID: "female-short-novel", Name: "女频短篇长篇小说转剧本", corpusID: "plugin_female_short_novel_to_script"},
		{ID: "default", Name: "中长篇纯对话", corpusID: "plugin_mid_length_dialogue"},
	},
	"ai-rewrite": {{ID: "default", Name: "短篇/长篇改文", body: "你是中文改文助手。保持原文事实、人物关系和核心含义不变，优化表达、节奏和可读性。只输出改写后的正文。"}},
	"storyboard-master": {
		{ID: "jimman-storyboard-1", Name: "姬漫·分镜脚本-推文分镜1", corpusID: "plugin_jimman_storyboard_tweet"},
		{ID: "default", Name: "【AI工具箱】分镜大师", corpusID: "plugin_ai_toolbox_storyboard"},
	},
	"character-edit": {{ID: "default", Name: "【AI工具箱】全文修改角色", corpusID: "plugin_ai_toolbox_character_rename"}},
	"viral-opening": {
		{ID: "short-drama", Name: "短剧爆款开头", corpusID: "plugin_short_drama_hook"},
		{ID: "default", Name: "超强爆款开头", corpusID: "plugin_strong_hook"},
	},
	"cover-generation": {{ID: "cover-design", Name: "一键封面指令", corpusID: "plugin_cover_design"}},
}

func pluginPromptBody(item PluginPromptOption) (string, bool) {
	if item.corpusID == "" {
		return item.body, item.body != ""
	}
	if body, ok := luoshuiOverrideBody(item.corpusID); ok {
		return body, true
	}
	registry, err := luoshui.Default()
	if err != nil {
		return "", false
	}
	return registry.Body(item.corpusID)
}

func PluginToolPromptOptions(toolID string) ([]PluginPromptOption, bool) {
	items, ok := pluginToolPrompts[strings.TrimSpace(toolID)]
	if !ok {
		return nil, false
	}
	result := make([]PluginPromptOption, len(items))
	for index, item := range items {
		result[index] = PluginPromptOption{ID: item.ID, Name: item.Name}
	}
	return result, true
}

func PluginToolSystemPrompt(toolID string, promptID ...string) (string, bool) {
	items, ok := pluginToolPrompts[strings.TrimSpace(toolID)]
	if !ok || len(items) == 0 {
		return "", false
	}
	wanted := "default"
	if len(promptID) > 0 && strings.TrimSpace(promptID[0]) != "" {
		wanted = strings.TrimSpace(promptID[0])
	}
	for _, item := range items {
		if item.ID == wanted {
			return pluginPromptBody(item)
		}
	}
	return "", false
}

// BuildCoverPrompt 把封面指令原文与本次短剧名、文案、字体和比例组装成完整指令。
// 「一键封面指令」要求模型只输出最终封面提示词正文，因此这里把用户输入和
// 比例要求作为「本次封面需求」追加在原文之后。
func BuildCoverPrompt(promptID, content, dramaName, aspectRatio, fontStyle string, referenceCount int) (string, bool) {
	if strings.TrimSpace(promptID) == "" {
		promptID = "cover-design"
	}
	system, ok := PluginToolSystemPrompt("cover-generation", promptID)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(system) + "\n\n" + BuildCoverRequirement(content, dramaName, fontStyle, aspectRatio, referenceCount), true
}

// BuildCoverRequirement 生成“本次封面需求”段，供直接组装和文本模型生成两条路径复用。
func BuildCoverRequirement(content, dramaName, fontStyle, aspectRatio string, referenceCount int) string {
	var builder strings.Builder
	builder.WriteString("## 本次封面需求\n")
	dramaName = strings.TrimSpace(dramaName)
	if dramaName == "" {
		builder.WriteString("短剧名：（未提供，请根据内容生成贴合番茄/七猫调性的书名）\n")
	} else {
		fmt.Fprintf(&builder, "短剧名：%s\n", dramaName)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		builder.WriteString("文案/简介：（未提供）\n")
	} else {
		fmt.Fprintf(&builder, "文案/简介：%s\n", content)
	}
	if fontStyle = strings.TrimSpace(fontStyle); fontStyle != "" {
		fmt.Fprintf(&builder, "字体设计：%s（以此为准，覆盖自动判断）\n", fontStyle)
	}
	if referenceCount > 0 {
		fmt.Fprintf(&builder, "参考图片：已提供 %d 张，请严格参考图片中的人物外貌。\n", referenceCount)
	} else {
		builder.WriteString("参考图片：未提供。\n")
	}
	ratio := strings.TrimSpace(aspectRatio)
	if ratio == "" {
		ratio = "3:4"
	}
	fmt.Fprintf(&builder, "比例：%s（请在最终提示词末尾保留 --ar %s）\n", ratio, ratio)
	return builder.String()
}
