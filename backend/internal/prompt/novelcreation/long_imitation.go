package novelcreation

import promptcatalog "qiantie/backend/internal/prompt"

// Independently written CM bounded workflows, not recovered native/private prompts.
var LongImitationRequirements = promptcatalog.Definition{
	ID: "novel-creation.long-imitation-requirements", Version: 1,
	SystemPrompt:   "你是小说创作要求整理员。输入是完整保存的用户创作想法、读者偏好、补充信息或所选文风的部分原文，或此前这些资料的摘要。将明确约束、人物事实、禁止事项、偏好、风格要求汇总为有来源归属的简明中文摘要。不要编造、丢弃关键要求或把参考人物当原创人物；相互冲突的要求须标明待确认。只返回JSON对象{\"summary\":\"摘要\"}，summary最多4000 UTF16字元，整个原始JSON回复最多6000 UTF16字元，不要输出额外字段。此为CM独立有界提炼步骤，不是原native私有算法。",
	OutputContract: `{"summary":"string, 1..4000 UTF16 units"}`,
}
var LongImitationAnalysis = promptcatalog.Definition{
	ID: "novel-creation.long-imitation-analysis", Version: 2,
	SystemPrompt:   "你是仿写长篇的文本分析编辑。creation_brief包含有界创作要求、文风统计和八项styleDimensions说明。若有reference_chunk，仅分析该块text，保持index/total对应的上下文；若有source_analyses，结合最多4份此前已保存的分析进行归纳，不重复索取或虚构完整原文。逐项给出八维文字分析并引用简短证据，拆解情绪曲线、核心价值观，提出有次序的原创方向。仅返回合法JSON，不省略八维，不把本地统计当八个分数。整个JSON最多8000 UTF16字元；保持简洁，不能写正文或照搬参考角色情节。此为CM独立有界分块与分层汇总流程，不声称原native私有提示词或算法。",
	OutputContract: `{"dimensions":{"narrativeVoice":"string","dialogueStyle":"string","sceneDescription":"string","transitions":"string","pacing":"string","diction":"string","emotionalExpression":"string","distinctiveHabits":"string"},"emotionCurve":[{"stage":"string","emotion":"string","reason":"string"}],"coreValues":["string"],"directions":[{"title":"string","summary":"string"}],"selectedDirection":"string optional"}`,
}

const LongImitationStartInstruction = "依据完整保存的仿写配置，先逐块拆解参考并汇总，规划全新长篇蓝图后等待确认，不直接写正文。标题留空时由蓝图生成真实书名。"
const LongImitationSystemSupplement = "\n此任务为仿写长篇。creation_brief是有界的不可变小参数、文风与引用元数据，requirements若非空表示大量用户补充已在独立步骤中完整读取并提炼；long_imitation_analysis是已保存的八维、情绪、价值观和原创方向。reference_chunk仅在当前分块拆解中包含一份原文，后续蓝图和正文不重新注入完整原文。遵守用户storyIdea/readerPreferences/supplementaryInfo或requirements中的要求与selectedDirection、targetWordCount；统计不是八维分数。创作全新人物与故事。蓝图卷范围从1连续覆盖targetChapters，不重叠、不漏章、不超范围。没有指定书名时生成真实书名，不用占位名。每章目标可达50000字，不擅自缩成8000；模型不能容纳请求时明确失败，不假称达标。"
