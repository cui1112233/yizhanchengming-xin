// Package novelcreation owns the system-prompt contracts for long-form novel
// creation.  Callers retain only a definition identity and provide the actual
// project data as user input.
package novelcreation

import promptcatalog "qiantie/backend/internal/prompt"

var FoundationPlan = promptcatalog.Definition{
	ID:           "novel-creation.foundation-plan",
	Version:      1,
	SystemPrompt: "你是长篇原创小说策划编辑。依据用户提供的创作要求，建立能持续写作的故事基础设定。只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。不得编造未提供的事实为既定资料；信息不足时在 JSON 中明确列出待确认内容。若输入包含 style_profile 或 rules，必须把它们当作硬性约束。",
	RequiredInputs: []string{
		"project_brief", "user_message",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"title":"string","genre":"string","world":"string","characters":[{"name":"string","role":"string","goal":"string","conflict":"string"}],"coreConflict":"string","volumes":[{"number":"integer","title":"string","goal":"string","chapterRange":["integer","integer"]}]}}`,
	Errors: map[string]string{
		"input_required":  "创作要求不能为空",
		"invalid_output":  "基础设定 JSON 格式无效",
		"upstream_failed": "基础设定生成失败",
	},
}

var ChapterPlan = promptcatalog.Definition{
	ID:           "novel-creation.chapter-plan",
	Version:      1,
	SystemPrompt: "你是长篇原创小说章节策划编辑。根据基础设定、已写内容和当前创作要求，安排一个可直接写作的章节。保持人物动机、世界规则和时间线一致，只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。style_profile 与 rules 是硬性约束。",
	RequiredInputs: []string{
		"foundation_plan", "chapter_number", "memory_context", "user_message",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"chapterNumber":"integer","title":"string","goal":"string","conflict":"string","turn":"string","hook":"string","targetWords":"integer"}}`,
	Errors: map[string]string{
		"context_required": "章节策划上下文不能为空",
		"invalid_output":   "章节策划 JSON 格式无效",
		"upstream_failed":  "章节策划生成失败",
	},
}

var ChapterDraft = promptcatalog.Definition{
	ID:           "novel-creation.chapter-draft",
	Version:      1,
	SystemPrompt: "你是长篇原创小说作者。根据章节计划和已确认记忆写出完整章节正文，保证事件、人物、时间线和伏笔连续；不要把计划或审核意见写入正文。只返回一个合法 JSON 对象，正文必须放在 content 字段，不要输出 Markdown 代码围栏、标题或解释。必须遵守 style_profile；rules 是不可违反的禁区。",
	RequiredInputs: []string{
		"foundation_plan", "chapter_plan", "memory_context",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"content":"string"}}`,
	Errors: map[string]string{
		"plan_required":   "章节计划不能为空",
		"invalid_output":  "章节正文 JSON 格式无效",
		"upstream_failed": "章节正文生成失败",
	},
}

var ChapterAudit = promptcatalog.Definition{
	ID:           "novel-creation.chapter-audit",
	Version:      1,
	SystemPrompt: "你是长篇原创小说质检编辑。检查章节正文是否符合章节目标，并核对人物、世界规则、时间线、事实和伏笔的一致性。每个问题必须有稳定 id、severity、description；可补充 category 和 evidence，且问题必须具体并可修复。rules 中的每一条都必须逐项核对。只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。",
	RequiredInputs: []string{
		"foundation_plan", "chapter_plan", "chapter_content", "memory_context",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"score":"number 0..100","issues":[{"id":"string","severity":"low|medium|high|critical","description":"string","category":"string optional","evidence":"string optional"}],"passed":"boolean"}}`,
	Errors: map[string]string{
		"content_required": "章节正文不能为空",
		"invalid_output":   "章节审核 JSON 格式无效",
		"upstream_failed":  "章节审核失败",
	},
}

var ChapterRevision = promptcatalog.Definition{
	ID:           "novel-creation.chapter-revision",
	Version:      1,
	SystemPrompt: "你是长篇原创小说修订编辑。只修复审核指出的问题，保留已经成立的剧情、人物声音和有效内容；不得为了修订凭空改写未被指出的关键事实。resolvedIssueIDs 只能引用审核 issues 中的 issue.id。只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。",
	RequiredInputs: []string{
		"chapter_content", "chapter_plan", "audit_issues", "memory_context",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"strategy":"string","content":"string","resolvedIssueIDs":["string issue.id"]}}`,
	Errors: map[string]string{
		"issues_required": "审核问题不能为空",
		"invalid_output":  "章节修订 JSON 格式无效",
		"upstream_failed": "章节修订失败",
	},
}

var TruthSettlement = promptcatalog.Definition{
	ID:           "novel-creation.truth-settlement",
	Version:      1,
	SystemPrompt: "你是长篇原创小说事实结算编辑。根据最终章节正文，提取已经发生的事实、人物变化、时间线变化和可追踪伏笔。只记录正文明确支持的内容，不把推测写成事实。只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。",
	RequiredInputs: []string{
		"chapter_content", "previous_truth", "chapter_number",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"facts":[{"id":"string","statement":"string"}],"characterChanges":[{"character":"string","change":"string"}],"timeline":[{"event":"string","position":"string"}],"foreshadowing":[{"thread":"string","status":"open|resolved"}]}}`,
	Errors: map[string]string{
		"content_required": "章节正文不能为空",
		"invalid_output":   "事实结算 JSON 格式无效",
		"upstream_failed":  "事实结算失败",
	},
}

var MemoryIndex = promptcatalog.Definition{
	ID:           "novel-creation.memory-index",
	Version:      1,
	SystemPrompt: "你是长篇原创小说记忆编辑。把本章最终结果整理成下一章可用的短记忆，保留关键事实、实体关系和未解决线索，不重复整段正文。只返回一个合法 JSON 对象，不要输出 Markdown、代码围栏、标题或解释。",
	RequiredInputs: []string{
		"chapter_bundle", "previous_memory", "truth_settlement",
	},
	OutputContract: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"summary":"string","entities":[{"name":"string","type":"string","state":"string"}],"openThreads":[{"id":"string","thread":"string","nextExpectation":"string"}]}}`,
	Errors: map[string]string{
		"bundle_required": "章节结果不能为空",
		"invalid_output":  "记忆索引 JSON 格式无效",
		"upstream_failed": "记忆索引生成失败",
	},
}

// All returns the complete, backend-owned prompt catalog in pipeline order.
func All() []promptcatalog.Definition {
	definitions := []promptcatalog.Definition{
		FoundationPlan,
		ChapterPlan,
		ChapterDraft,
		LengthNormalize, PostWrite, AiTells,
		AgentChapterAudit,
		ChapterRevision,
		TruthSettlement,
		MemoryIndex,
	}
	for index := range definitions {
		definitions[index] = cloneDefinition(definitions[index])
	}
	return definitions
}

func cloneDefinition(definition promptcatalog.Definition) promptcatalog.Definition {
	definition.RequiredInputs = append([]string(nil), definition.RequiredInputs...)
	errors := make(map[string]string, len(definition.Errors))
	for code, message := range definition.Errors {
		errors[code] = message
	}
	definition.Errors = errors
	return definition
}

const StructureRepairInstruction = "上一轮输出不符合合同，只返回修复后的完整 JSON 对象"
const StructureRepairSystem = "上一轮输出无法通过合同校验。现在只返回一个修复后的完整 JSON 对象，不要解释。"

var LengthNormalize = promptcatalog.Definition{ID: "novel-creation.length-normalize", Version: 1, SystemPrompt: "你是章节长度整理编辑。依据目标字数适度扩充或精简正文，保留事件、人物动机和段落结尾完整，不截断句子，不改变已经发生的事实。只返回包含完整content的JSON对象。", RequiredInputs: []string{"chapter_content", "target_chapter_words"}, OutputContract: ChapterDraft.OutputContract, Errors: map[string]string{"content_required": "章节正文不能为空", "invalid_output": "正文整理 JSON 格式无效", "upstream_failed": "正文整理失败"}}
var PostWrite = promptcatalog.Definition{ID: "novel-creation.post-write", Version: 1, SystemPrompt: "你是小说正文后处理编辑。修正错字、标点、段落衔接和叙述一致性，保持人物声音、事件顺序与事实，不插入解释。只返回包含完整content的JSON对象。", RequiredInputs: []string{"chapter_content"}, OutputContract: ChapterDraft.OutputContract, Errors: map[string]string{"content_required": "章节正文不能为空", "invalid_output": "正文整理 JSON 格式无效", "upstream_failed": "正文整理失败"}}
var AiTells = promptcatalog.Definition{ID: "novel-creation.ai-tells", Version: 1, SystemPrompt: "你是小说表达检查编辑。检查并整理机械重复、空泛总结、模板化转折和不自然的解释句，保留具体行动、个性语言、剧情事实及完整结尾。只返回包含完整content的JSON对象。", RequiredInputs: []string{"chapter_content"}, OutputContract: ChapterDraft.OutputContract, Errors: map[string]string{"content_required": "章节正文不能为空", "invalid_output": "正文整理 JSON 格式无效", "upstream_failed": "正文整理失败"}}

// Preserve ChapterAudit v1 for already-journaled legacy requests.
var AgentChapterAudit = func() promptcatalog.Definition {
	d := cloneDefinition(ChapterAudit)
	d.Version = 2
	d.SystemPrompt += " 输入中的pass_score是本次审核通过分。passed必须等于score达到pass_score且issues中不存在high或critical问题；不得使用其他固定通过分。"
	return d
}()
