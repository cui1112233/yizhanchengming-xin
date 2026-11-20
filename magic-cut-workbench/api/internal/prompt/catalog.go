package prompt

import "sort"

type Template struct {
	ID                string   `json:"id"`
	Purpose           string   `json:"purpose"`
	System            string   `json:"system"`
	RequiredVariables []string `json:"requiredVariables"`
	SourceAuthority   string   `json:"sourceAuthority"`
	SourceLocation    string   `json:"sourceLocation"`
}

type Catalog struct{}

var templates = map[string]Template{
	"magic_cut.ai_pre_roll.negative_constraints.v1": {
		ID:                "magic_cut.ai_pre_roll.negative_constraints.v1",
		Purpose:           "AI 前贴视频默认负面约束",
		System:            "不生成字幕；画面无水印、台标和乱码。",
		RequiredVariables: []string{},
		SourceAuthority:   "observed-client",
		SourceLocation:    "app.asar/assets/index-lgtcjTmg.js:16",
	},
	"magic_cut.ai_pre_roll.narration_mix.v1": {
		ID:                "magic_cut.ai_pre_roll.narration_mix.v1",
		Purpose:           "AI 前贴独立旁白与画面声音的生成边界",
		System:            "解说旁白由独立配音音轨合成，不生成解说旁白；人物对白和环境声按画面提示词生成。",
		RequiredVariables: []string{},
		SourceAuthority:   "observed-client",
		SourceLocation:    "app.asar/assets/index-lgtcjTmg.js:165",
	},
	"magic_cut.hot_replication.infer_video.v1": {
		ID:      "magic_cut.hot_replication.infer_video.v1",
		Purpose: "把爆款视频理解结果整理成可编辑的生成提示词",
		System: "你是视频复刻分析助手。请根据提供的视频理解结果，提取主体、场景、动作、镜头、光线、节奏和声音，输出可直接用于视频生成的完整中文提示词。" +
			"不要解释，不要添加输入中不存在的品牌或人物身份。\n\n视频理解：{{video_summary}}\n复刻模式：{{mode}}",
		RequiredVariables: []string{"video_summary", "mode"},
		SourceAuthority:   "derived-equivalent",
		SourceLocation:    "behavior: /hot-replication AI 推理",
	},
	"magic_cut.script_batch.match_shots.v1": {
		ID:      "magic_cut.script_batch.match_shots.v1",
		Purpose: "把脚本分镜与已导入素材做可解释的匹配",
		System: "你是短视频分镜匹配助手。按脚本顺序逐镜选择最合适的素材，优先匹配人物、场景、动作和情绪，不得使用素材清单之外的编号。" +
			"只输出结构化匹配结果和简短原因。\n\n分镜：{{shots}}\n素材清单：{{material_catalog}}",
		RequiredVariables: []string{"shots", "material_catalog"},
		SourceAuthority:   "derived-equivalent",
		SourceLocation:    "behavior: /script-batch-compose/:id? 画面匹配",
	},
}

func (Catalog) Get(id string) (Template, bool) {
	template, ok := templates[id]
	return template, ok
}

func (Catalog) SystemFingerprints() []string {
	ids := make([]string, 0, len(templates))
	for id := range templates {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, templates[id].System)
	}
	return result
}
