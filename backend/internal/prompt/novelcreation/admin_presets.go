package novelcreation

import catalog "qiantie/backend/internal/prompt"

const AdminModule = "novel-creation"
const NativeRevisionPresetID = "novel-creation.native-revision"
const PatchOnlyPresetID = "novel-creation.revision-patch-only"
const RewriteOnlyPresetID = "novel-creation.revision-rewrite-only"
const IncrementalSettlementPresetID = "novel-creation.settlement-incremental"
const ImitationSupplementPresetID = "novel-creation.imitation-supplement"
const StudioPlanningPresetID = "novel-creation.studio-planning"

type AdminPrompt struct {
	ID, Name, Slot string
	Definition     catalog.Definition
}

// Runtime and administrator defaults share these definitions; no frontend copy.
func AdminPrompts() []AdminPrompt {
	return []AdminPrompt{
		{FoundationPlan.ID, "长篇创作蓝图", "novel.long", FoundationPlan},
		{ChapterPlan.ID, "章节规划", "novel.long", ChapterPlan},
		{ChapterDraft.ID, "章节正文", "novel.long", ChapterDraft},
		{LengthNormalize.ID, "章节字数整理", "novel.long", LengthNormalize},
		{AgentChapterAudit.ID, "章节审稿", "novel.review", AgentChapterAudit},
		{NativeRevisionPresetID, "局部修改与整章修订：基础提示词", "novel.review", NativeChapterRevision},
		{PatchOnlyPresetID, "局部修改：分流指令", "novel.review", catalog.Definition{SystemPrompt: nativePatchOnlyInstruction}},
		{RewriteOnlyPresetID, "整章重写：分流指令", "novel.review", catalog.Definition{SystemPrompt: nativeRewriteOnlyInstruction}},
		{TruthSettlement.ID, "事实结算（资料重建）", "novel.state", TruthSettlement},
		{IncrementalSettlementPresetID, "事实与伏笔增量结算", "novel.state", IncrementalTruthSettlement},
		{MemoryIndex.ID, "下一章记忆索引", "novel.state", MemoryIndex},
		{LongImitationRequirements.ID, "长篇仿写要求整理", "novel.imitation", LongImitationRequirements},
		{LongImitationAnalysis.ID, "长篇仿写参考分析", "novel.imitation", LongImitationAnalysis},
		{ImitationSupplementPresetID, "长篇仿写共用补充规则", "novel.imitation", catalog.Definition{SystemPrompt: LongImitationSystemSupplement}},
		{ShortSourceAnalysis.ID, "短篇仿写参考分析", "novel.short", ShortSourceAnalysis},
		{ShortOutlinePrompt.ID, "短篇大纲", "novel.short", ShortOutlinePrompt},
		{ShortReviewPrompt.ID, "短篇审稿", "novel.short", ShortReviewPrompt},
		{ShortOutlineRevision.ID, "短篇大纲修订", "novel.short", ShortOutlineRevision},
		{ShortBodyPrompt.ID, "短篇正文", "novel.short", ShortBodyPrompt},
		{ShortContinuity.ID, "短篇连续性摘要", "novel.short", ShortContinuity},
		{ShortBodyRevision.ID, "短篇正文修订", "novel.short", ShortBodyRevision},
		{ShortRecommendation.ID, "短篇推荐文案", "novel.short", ShortRecommendation},
		{StyleAnalysis.ID, "写作文风分析", "novel.tools", StyleAnalysis},
		{StudioPlanningPresetID, "创作对话与建书建议", "novel.tools", catalog.Definition{SystemPrompt: StudioPlanningSystem}},
	}
}

func AdminSlots() []map[string]string {
	return []map[string]string{
		{"id": "novel.long", "label": "长篇蓝图与正文", "mode": "primary"},
		{"id": "novel.review", "label": "审稿与修订", "mode": "primary"},
		{"id": "novel.state", "label": "事实、伏笔与记忆", "mode": "primary"},
		{"id": "novel.imitation", "label": "长篇仿写", "mode": "primary"},
		{"id": "novel.short", "label": "短篇创作", "mode": "primary"},
		{"id": "novel.tools", "label": "创作对话与文风", "mode": "primary"},
	}
}

// Startup-only seeding through the existing durable admin writer. Existing
// IDs, including archived records and unpublished drafts, are never overwritten.
func EnsureAdminPresets(path string) error {
	store := catalog.NewAdminStore(path)
	rows, e := store.List(AdminModule)
	if e != nil {
		return e
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if id, ok := r["id"].(string); ok {
			seen[id] = true
		}
	}
	for _, p := range AdminPrompts() {
		if seen[p.ID] {
			continue
		}
		draft, e := store.CreateDraft("system", map[string]any{"id": p.ID, "module": AdminModule, "name": p.Name, "kind": "base", "description": "对应小说创作实际调用。发布影响新任务，已创建任务保留原版本；输出结构由后端固定。", "body": p.Definition.SystemPrompt, "compatibleBaseIds": []string{}, "protocolLock": map[string]any{"slot": p.Slot}})
		if e != nil {
			return e
		}
		if _, e = store.Publish("system", p.ID, draft["version"].(int)); e != nil {
			return e
		}
	}
	return nil
}
