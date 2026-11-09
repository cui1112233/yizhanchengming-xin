package asset

import promptcatalog "qiantie/backend/internal/prompt"

var ContextReference = promptcatalog.Definition{
	ID: "asset.context-reference", Version: 1,
	SystemPrompt:   "你是项目素材上下文整理器。只使用后端提供的稳定素材和节点 ID，按原顺序整理用户明确选择的引用；不得补入已取消选择、跨项目或不存在的内容。",
	RequiredInputs: []string{"selected_ids", "project_assets", "canvas_nodes"},
	OutputContract: "返回去重后的稳定 ID 列表以及每项的来源类型。",
	Errors:         map[string]string{"reference_missing": "引用对象不存在", "reference_out_of_scope": "引用对象不属于当前项目"},
}

var VisualReference = promptcatalog.Definition{
	ID: "asset.visual-reference", Version: 1,
	SystemPrompt:   "你是视觉参考整理器。根据人物、场景和道具素材生成简洁的参考说明，保持素材身份和版本，不把参考图描述成已经生成的结果。",
	RequiredInputs: []string{"asset_ids", "asset_versions", "generation_kind"},
	OutputContract: "返回按输入顺序排列的参考说明和稳定素材 ID。",
	Errors:         map[string]string{"unsupported_kind": "当前生成类型不支持该素材", "asset_unavailable": "素材暂时不可用"},
}
