package storyboard

import promptcatalog "qiantie/backend/internal/prompt"

var ScriptSplit = promptcatalog.Definition{
	ID: "storyboard.script-split", Version: 1,
	SystemPrompt:   "把剧本文本拆成连续、可生产的结构化分镜。每段必须保留时长、场景、人物、对白或旁白、图片提示词和视频镜头时间线；不得省略原文关键事件。",
	RequiredInputs: []string{"script_node_id", "source_text", "max_video_duration"},
	OutputContract: "返回符合导演 JSON 契约的 storyboard 数组，每项包含 duration_sec、visual_prompt、shots 和 video_desc。",
	Errors:         map[string]string{"source_empty": "剧本正文为空", "invalid_output": "分镜结果格式无效"},
}

var ShotMerge = promptcatalog.Definition{
	ID: "storyboard.shot-merge", Version: 1,
	SystemPrompt:   "按原顺序合并相邻分镜，不改写对白事实，不丢失提示词、素材引用或历史生成结果。",
	RequiredInputs: []string{"shots", "target_duration_seconds"},
	OutputContract: "目标时长只能为 10、15 或 30 秒；输出稳定新镜头 ID、来源镜头 ID、合并时长和顺序拼接字段。",
	Errors:         map[string]string{"invalid_duration": "合并时长无效", "non_adjacent": "只能合并相邻镜头"},
}
