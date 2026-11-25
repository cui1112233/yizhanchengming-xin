package novelcreation

import promptcatalog "qiantie/backend/internal/prompt"

// 洛水 novels.pipeline.play_runner + play_agents 的提示词合同。
// 为控制每次推进的模型调用次数，CM 把 parse_action 与 simulate 合并为一次
// PlaySimulate；narrate 与 extract_state 保持独立，语义与原文一致。
var PlayOpening = promptcatalog.Definition{
	ID:      "novel-creation.play-opening",
	Version: 1,
	SystemPrompt: "你是互动叙事开场编剧。输出纯叙事开场，不要写选项列表。" +
		"只返回一个合法 JSON 对象，字段 sceneText 为开场正文，suggestedActions 为 2-4 个短动作建议。" +
		"不要输出 Markdown 代码围栏或解释。",
	RequiredInputs: []string{"world_title", "premise", "mode"},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","structure":{"sceneText":"string","suggestedActions":["string"]}}`,
	Errors: map[string]string{
		"premise_required": "互动世界 premise 不能为空",
		"invalid_output":   "开场 JSON 格式无效",
		"upstream_failed":  "开场生成失败",
	},
}

var PlaySimulate = promptcatalog.Definition{
	ID:      "novel-creation.play-simulate",
	Version: 1,
	SystemPrompt: "你是互动叙事的世界推演引擎。先判断玩家行动的类型与意图，再推演这一拍在世界里造成什么后果。" +
		"只写已发生的事实，不替玩家做决定，不写正文描写。封锁的行动要说明原因。" +
		"只返回一个合法 JSON 对象，不要输出 Markdown 代码围栏或解释。",
	RequiredInputs: []string{"world_premise", "state_summary", "previous_scene", "player_input", "turn"},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","structure":{"actionKind":"look|say|move|do|wait","intent":"string","outcome":"string","location":"string","clock":"string","blocked":"boolean","blockedReason":"string","entityChanges":[{"name":"string","change":"string"}],"relationChanges":[{"from":"string","to":"string","type":"string","value":"string"}],"newFacts":["string"],"goals":["string"]}}`,
	Errors: map[string]string{
		"input_required":  "行动输入不能为空",
		"invalid_output":  "世界推演 JSON 格式无效",
		"upstream_failed": "世界推演失败",
	},
}

var PlayNarrate = promptcatalog.Definition{
	ID:      "novel-creation.play-narrate",
	Version: 1,
	SystemPrompt: "你是互动叙事的场景渲染者。把推演结果写成一幕可读的场景，保持人称与世界观一致，" +
		"不预告后续选项，不写分析术语。只返回一个合法 JSON 对象，sceneText 为正文，suggestedActions 为 2-4 个短动作建议。" +
		"不要输出 Markdown 代码围栏或解释。",
	RequiredInputs: []string{"world_premise", "state_summary", "previous_scene", "player_input", "outcome"},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","structure":{"sceneText":"string","suggestedActions":["string"]}}`,
	Errors: map[string]string{
		"input_required":  "场景渲染输入不能为空",
		"invalid_output":  "场景渲染 JSON 格式无效",
		"upstream_failed": "场景渲染失败",
	},
}

var PlayStateUpdate = promptcatalog.Definition{
	ID:      "novel-creation.play-state-update",
	Version: 1,
	SystemPrompt: "你是互动叙事的状态记录员。根据这一幕正文与推演结果，更新世界状态中真正发生变化的字段：" +
		"地点、时间、实体、关系、事实与目标。没有变化的字段留空数组或原值，不要编造未发生的内容。" +
		"只返回一个合法 JSON 对象，不要输出 Markdown 代码围栏或解释。",
	RequiredInputs: []string{"scene_text", "previous_state", "turn"},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","structure":{"location":"string","clock":"string","entities":[{"name":"string","type":"string","status":"string","location":"string"}],"relations":[{"from":"string","to":"string","type":"string","value":"string"}],"facts":["string"],"goals":["string"]}}`,
	Errors: map[string]string{
		"scene_required":  "场景正文不能为空",
		"invalid_output":  "状态更新 JSON 格式无效",
		"upstream_failed": "状态更新失败",
	},
}

var PlayDefinitions = []promptcatalog.Definition{PlayOpening, PlaySimulate, PlayNarrate, PlayStateUpdate}
