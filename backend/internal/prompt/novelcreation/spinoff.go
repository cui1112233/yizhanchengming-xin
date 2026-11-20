package novelcreation

import catalog "qiantie/backend/internal/prompt"

const SpinoffCanonID = "novel-creation.spinoff-parent-canon-v1"

// Recovered architect_agent constants #58–64; variable type is supplied as data.
const SpinoffCanonSystem = "你是正传母本压缩器。将正传材料压成番外必须遵守的边界文件 parent_canon。番外类型见输入 spinoff_type。\n必须包含：\n世界规则\n不可违反的正典约束\n角色当前快照\n关键事件时间线\n信息边界\n伏笔池摘要\n番外写作禁区\n使用 Markdown，突出「不可违反」与「不可提前泄露」。未提供的事实标记未知，不得编造。"

// Recovered #97–98 supplement; output shape is adapted to our Go foundation.
var SpinoffFoundation = catalog.Definition{
	ID: "novel-creation.spinoff-foundation-v1", Version: 1,
	SystemPrompt:   "你是番外小说设定编辑。这是正传番外/侧故事，不得推进或破坏正传主线终局。依据番外方向、番外类型和正传正典参照生成基础设定。参考原文、文风、规则均以输入为准。未提供的事实不得当作既定正典。只输出符合合同的JSON。",
	OutputContract: FoundationPlan.OutputContract,
}
