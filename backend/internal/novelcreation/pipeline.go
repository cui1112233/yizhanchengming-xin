package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"qiantie/backend/internal/batchfactoryv11"
	promptcatalog "qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

const (
	ErrInvalidOutput = "novel_invalid_output"
	ErrRunCancelled  = "novel_run_cancelled"

	defaultAuditThreshold = 85.0
	defaultMaxRevisions   = 2
)

// ProviderResolver resolves a member's configured model without exposing
// credentials to the pipeline or to the browser.
type ProviderResolver func(ctx context.Context, ownerUsername, modelID string) (batchfactoryv11.DirectorProvider, error)

// EventSink preserves the durable-first event rule. Persist must return the
// durable event identity; only that returned event is sent to Publish.
type EventSink interface {
	Persist(context.Context, RunEvent) (RunEvent, error)
	Publish(context.Context, RunEvent) error
}

// RunContext is the immutable input snapshot for one long-form chapter run.
// Runtime is used only for cancellation checks; durable writes remain the
// caller's responsibility after Run returns a complete ChapterBundle.
type RunContext struct {
	LongImitation         *LongImitationCreation
	ImitationAnalysis     *LongImitationAnalysis
	ImitationRequirements *LongImitationRequirements
	OwnerUsername         string
	Run                   Run
	Project               Project
	Session               Session
	UserMessage           string
	ChapterNumber         int
	PreviousChapters      []PreviousChapter
	PreviousTruth         json.RawMessage
	PreviousMemory        json.RawMessage
	Runtime               Runtime
	// ModeInputs is the validated mode-specific contract payload.
	ModeInputs map[string]string
	// StyleProfile and Rules are project-level writing constraints injected into
	// every generation step.
	StyleProfile string
	Rules        []string
	// Checkpoint pauses the run after the blueprint so the user can confirm or
	// edit it before any chapter body is written.
	Checkpoint bool
	// ConfirmedFoundation skips blueprint generation and reuses an existing or
	// user-confirmed blueprint.
	ConfirmedFoundation json.RawMessage
	// KnowledgeMaterials is the bounded snapshot of user-selected materials.
	Notes              []ProjectNote
	KnowledgeMaterials []KnowledgeMaterial
	// MaxRevisionRounds overrides the default audit-revision loop bound.
	MaxRevisionRounds int
	Agent             *AgentSettings
}

type Pipeline struct {
	resolveProvider ProviderResolver
	auditThreshold  float64
	maxRevisions    int
}

func NewPipeline(resolveProvider ProviderResolver) *Pipeline {
	return &Pipeline{
		resolveProvider: resolveProvider,
		auditThreshold:  defaultAuditThreshold,
		maxRevisions:    defaultMaxRevisions,
	}
}

type foundationPlanOutput struct {
	Title        string                `json:"title"`
	Genre        string                `json:"genre"`
	World        string                `json:"world"`
	Characters   []foundationCharacter `json:"characters"`
	CoreConflict string                `json:"coreConflict"`
	Volumes      []foundationVolume    `json:"volumes"`
}

type foundationCharacter struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Goal     string `json:"goal"`
	Conflict string `json:"conflict"`
}

type foundationVolume struct {
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Goal         string `json:"goal"`
	ChapterRange []int  `json:"chapterRange"`
}

type chapterPlanOutput struct {
	ChapterNumber int    `json:"chapterNumber"`
	Title         string `json:"title"`
	Goal          string `json:"goal"`
	Conflict      string `json:"conflict"`
	Turn          string `json:"turn"`
	Hook          string `json:"hook"`
	TargetWords   int    `json:"targetWords"`
}

type chapterDraftOutput struct {
	Content string `json:"content"`
}

type chapterAuditOutput struct {
	Score  *float64     `json:"score"`
	Issues []auditIssue `json:"issues"`
	Passed *bool        `json:"passed"`
}

type auditIssue struct {
	ID          string          `json:"id"`
	Severity    string          `json:"severity"`
	Description string          `json:"description"`
	Category    json.RawMessage `json:"category,omitempty"`
	Evidence    json.RawMessage `json:"evidence,omitempty"`
}

type chapterRevisionOutput struct {
	Strategy         string   `json:"strategy"`
	Content          string   `json:"content"`
	ResolvedIssueIDs []string `json:"resolvedIssueIDs"`
}

type truthSettlementOutput struct {
	Facts            []settledFact            `json:"facts"`
	CharacterChanges []settledCharacterChange `json:"characterChanges"`
	Timeline         []settledTimelineEvent   `json:"timeline"`
	Foreshadowing    []settledForeshadowing   `json:"foreshadowing"`
}

type settledFact struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
}

type settledCharacterChange struct {
	Character string `json:"character"`
	Change    string `json:"change"`
}

type settledTimelineEvent struct {
	Event    string `json:"event"`
	Position string `json:"position"`
}

type settledForeshadowing struct {
	Thread string `json:"thread"`
	Status string `json:"status"`
}

type memoryIndexOutput struct {
	Summary     string             `json:"summary"`
	Entities    []memoryEntity     `json:"entities"`
	OpenThreads []memoryOpenThread `json:"openThreads"`
}

type memoryEntity struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	State string `json:"state"`
}

type memoryOpenThread struct {
	ID              string `json:"id"`
	Thread          string `json:"thread"`
	NextExpectation string `json:"nextExpectation"`
}

type modelCallSpec struct {
	key         string
	definition  promptcatalog.Definition
	step        RunStatus
	userPayload any
	temperature float64
	maxTokens   int
	validate    func() error
}

func (p *Pipeline) Run(ctx context.Context, input RunContext, sink EventSink) (ChapterBundle, error) {
	if err := p.validateRunContext(input, sink); err != nil {
		return ChapterBundle{}, err
	}
	if err := p.checkCancelled(ctx, input); err != nil {
		return ChapterBundle{}, err
	}
	if input.LongImitation != nil && input.ChapterNumber > input.LongImitation.Input.TargetChapters {
		return ChapterBundle{}, &DomainError{Code: ErrInvalidTransition, Message: "已达到目标章数"}
	}
	if input.LongImitation != nil {
		input.ModeInputs = nil
	}
	requirements, requirementsErr := p.imitationRequirements(ctx, input, sink)
	if requirementsErr != nil {
		return ChapterBundle{}, requirementsErr
	}
	input.ImitationRequirements = requirements
	if input.LongImitation != nil && input.ImitationAnalysis == nil && len(input.ConfirmedFoundation) == 0 {
		if e := p.emitStage(ctx, input, sink, RunPlanning, "正在拆解参考文风、情绪曲线与原创方向", 5); e != nil {
			return ChapterBundle{}, e
		}
	}
	analysis, analysisErr := p.imitationAnalysis(ctx, input, sink)
	if analysisErr != nil {
		return ChapterBundle{}, analysisErr
	}
	input.ImitationAnalysis = analysis

	var provider batchfactoryv11.DirectorProvider
	var err error

	if err := p.emitStage(ctx, input, sink, RunPlanning, "正在生成基础设定和章节计划", 10); err != nil {
		return ChapterBundle{}, err
	}
	foundationPayload := map[string]any{
		"creation_mode":       modeContract(input.Session.Mode),
		"mode_inputs":         input.ModeInputs,
		"style_profile":       input.StyleProfile,
		"rules":               input.Rules,
		"notes":               input.Notes,
		"knowledge_materials": input.KnowledgeMaterials,
		"project_brief": map[string]any{
			"title": input.Project.Title, "genre": input.Project.Genre,
			"world": input.Project.World, "protagonist": input.Project.Protagonist,
			"coreConflict":       input.Project.CoreConflict,
			"targetChapterWords": input.Project.TargetChapterWords,
		},
		"user_message": input.UserMessage,
	}
	if input.LongImitation != nil {
		foundationPayload["creation_brief"] = imitationModelBrief(input)
		foundationPayload["long_imitation_analysis"] = analysis
	}
	var foundation foundationPlanOutput
	var foundationRaw json.RawMessage
	if len(input.ConfirmedFoundation) > 0 {
		if err := decodeAndValidate(string(input.ConfirmedFoundation), &foundation, func() error {
			return validateImitationFoundation(foundation, input.LongImitation)
		}); err != nil {
			return ChapterBundle{}, &DomainError{Code: ErrInvalidOutput, Message: "confirmed foundation is invalid", Cause: err}
		}
		foundationRaw = cloneRawJSON(input.ConfirmedFoundation)
	} else {
		foundationRaw, err = p.callJSON(ctx, input, sink, provider, modelCallSpec{
			definition: novelprompt.FoundationPlan, step: RunPlanning,
			userPayload: foundationPayload, temperature: 0.4, maxTokens: 6000,
			validate: func() error { return validateImitationFoundation(foundation, input.LongImitation) },
		}, &foundation)
		if err != nil {
			return ChapterBundle{}, err
		}
		if input.Checkpoint {
			return checkpointBundle(input, foundationRaw)
		}
	}

	chapterNumber := input.ChapterNumber
	if chapterNumber <= 0 {
		chapterNumber = 1
	}
	var plan chapterPlanOutput
	planRaw, err := p.callJSON(ctx, input, sink, provider, modelCallSpec{
		definition: novelprompt.ChapterPlan, step: RunPlanning,
		userPayload: map[string]any{
			"creation_mode":        modeContract(input.Session.Mode),
			"mode_inputs":          input.ModeInputs,
			"style_profile":        input.StyleProfile,
			"rules":                input.Rules,
			"notes":                input.Notes,
			"knowledge_materials":  input.KnowledgeMaterials,
			"foundation_plan":      foundation,
			"chapter_number":       chapterNumber,
			"target_chapter_words": input.Project.TargetChapterWords,
			"previous_truth":       rawJSONValue(input.PreviousTruth),
			"previous_chapters":    input.PreviousChapters,
			"memory_context":       rawJSONValue(input.PreviousMemory),
			"user_message":         input.UserMessage,
		},
		temperature: 0.35, maxTokens: 3500,
		validate: func() error {
			return validateImitationPlan(plan, chapterNumber, foundation.Volumes, input.LongImitation)
		},
	}, &plan)
	if err != nil {
		return ChapterBundle{}, err
	}
	selectedVolume, err := outlineVolumeForChapter(foundation.Volumes, plan.ChapterNumber)
	if err != nil {
		return ChapterBundle{}, invalidOutputError(novelprompt.ChapterPlan, err)
	}

	if err := p.emitStage(ctx, input, sink, RunWriting, "正在写作章节正文", 30); err != nil {
		return ChapterBundle{}, err
	}
	var draft chapterDraftOutput
	draftRaw, err := p.callJSON(ctx, input, sink, provider, modelCallSpec{
		definition: novelprompt.ChapterDraft, step: RunWriting,
		userPayload: map[string]any{
			"creation_mode":       modeContract(input.Session.Mode),
			"mode_inputs":         input.ModeInputs,
			"style_profile":       input.StyleProfile,
			"rules":               input.Rules,
			"notes":               input.Notes,
			"knowledge_materials": input.KnowledgeMaterials,
			"foundation_plan":     foundation,
			"chapter_plan":        plan,
			"previous_truth":      rawJSONValue(input.PreviousTruth),
			"previous_chapters":   input.PreviousChapters,
			"memory_context":      rawJSONValue(input.PreviousMemory),
		},
		temperature: 0.75, maxTokens: 12000,
		validate: func() error { return requireText("content", draft.Content) },
	}, &draft)
	if err != nil {
		return ChapterBundle{}, err
	}

	if err := p.emitStage(ctx, input, sink, RunNormalizing, "正在规范化章节正文", 45); err != nil {
		return ChapterBundle{}, err
	}
	draft.Content = strings.TrimSpace(draft.Content)
	bestContent := draft.Content
	bestContentRaw := draftRaw
	if input.Agent != nil && input.Agent.EnableLengthNorm && needsLengthNormalization(bestContent, input.Project.TargetChapterWords) {
		var processed chapterDraftOutput
		processedRaw, callErr := p.callJSON(ctx, input, sink, provider, modelCallSpec{definition: novelprompt.LengthNormalize, step: RunNormalizing, userPayload: map[string]any{"chapter_content": bestContent, "chapter_plan": plan, "target_chapter_words": input.Project.TargetChapterWords, "rules": input.Rules, "style_profile": input.StyleProfile}, maxTokens: 12000, validate: func() error { return requireText("content", processed.Content) }}, &processed)
		if callErr != nil {
			return ChapterBundle{}, callErr
		}
		bestContent = strings.TrimSpace(processed.Content)
		bestContentRaw = processedRaw
	}

	if err := p.emitStage(ctx, input, sink, RunAuditing, "正在审核章节", 55); err != nil {
		return ChapterBundle{}, err
	}
	audits := make([]Audit, 0, p.maxRevisions+1)
	revisions := make([]Revision, 0, p.maxRevisions)
	currentAudit, currentAuditRaw, err := p.auditChapter(ctx, input, sink, provider, foundation, plan, bestContent, 0)
	if err != nil {
		return ChapterBundle{}, err
	}
	audits = append(audits, newAudit(input, len(audits)+1, currentAudit, currentAuditRaw))
	bestAudit := currentAudit
	bestAuditRaw := currentAuditRaw

	threshold := p.threshold(input)
	revisingStageEmitted := false
	maxRevisions := p.maxRevisions
	if input.Agent != nil {
		maxRevisions = input.Agent.MaxRevisionRounds
	} else if input.MaxRevisionRounds > 0 {
		maxRevisions = input.MaxRevisionRounds
	}
	for round := 1; !auditPasses(bestAudit, threshold) && round <= maxRevisions; round++ {
		if len(bestAudit.Issues) == 0 {
			break
		}
		if !revisingStageEmitted {
			if err := p.emitStage(ctx, input, sink, RunRevising, "正在根据审核修订章节", 65); err != nil {
				return ChapterBundle{}, err
			}
			revisingStageEmitted = true
		}

		var revision chapterRevisionOutput
		revisionRaw, callErr := p.callJSON(ctx, input, sink, provider, modelCallSpec{
			definition: novelprompt.ChapterRevision, step: RunRevising, key: fmt.Sprintf(":%d", round),
			userPayload: withReviewDimensionContext(input, map[string]any{
				"chapter_content":   bestContent,
				"chapter_plan":      plan,
				"audit_issues":      bestAudit.Issues,
				"previous_truth":    rawJSONValue(input.PreviousTruth),
				"previous_chapters": input.PreviousChapters,
				"memory_context":    rawJSONValue(input.PreviousMemory),
			}, foundation),
			temperature: 0.55, maxTokens: 12000,
			validate: func() error { return validateRevision(revision, bestAudit.Issues) },
		}, &revision)
		if callErr != nil {
			return ChapterBundle{}, callErr
		}
		revisions = append(revisions, newRevision(input, round, revision, revisionRaw))

		candidateAudit, candidateAuditRaw, auditErr := p.auditChapter(ctx, input, sink, provider, foundation, plan, revision.Content, round)
		if auditErr != nil {
			return ChapterBundle{}, auditErr
		}
		audits = append(audits, newAudit(input, len(audits)+1, candidateAudit, candidateAuditRaw))
		if betterAudit(candidateAudit, bestAudit, threshold) {
			bestContent = strings.TrimSpace(revision.Content)
			bestContentRaw = revisionRaw
			bestAudit = candidateAudit
			bestAuditRaw = candidateAuditRaw
		}
	}

	finalStatus := RunSucceeded
	if !auditPasses(bestAudit, threshold) {
		finalStatus = RunSucceededWithWarnings
	}

	var localChecks json.RawMessage
	if input.Agent != nil && (input.Agent.EnablePostWrite || input.Agent.EnableAiTells) {
		checks := map[string]any{"model_calls": 0, "content_rewritten": false}
		if input.Agent.EnablePostWrite {
			titles := []string{}
			// Full previous chapter texts are not currently present in RunContext.
			// Do not feed bounded endings to the full cross-chapter check.
			result := ValidatePostWrite(bestContent, PostWriteInput{Title: plan.Title, ExistingTitles: titles})
			checks["post_write"] = result
			checks["cross_chapter_status"] = "not_checked_missing_full_text"
			if !result.Passed {
				finalStatus = RunSucceededWithWarnings
			}
		}
		if input.Agent.EnableAiTells {
			checks["expression_metrics"] = localExpressionMetrics(bestContent)
		}
		payload, marshalErr := json.Marshal(checks)
		if marshalErr != nil {
			return ChapterBundle{}, marshalErr
		}
		localChecks = payload
		if err := p.emit(ctx, input, sink, RunEvent{Type: "run.local_checks", Step: string(RunNormalizing), Status: RunNormalizing, Message: "本地文字检查完成，未调用模型、未改写正文", Progress: 75, Payload: payload}); err != nil {
			return ChapterBundle{}, err
		}
	}

	if err := p.emitStage(ctx, input, sink, RunSettling, "正在结算事实、人物变化和伏笔", 80); err != nil {
		return ChapterBundle{}, err
	}
	var settlement truthSettlementOutput
	settlementRaw, err := p.callJSON(ctx, input, sink, provider, modelCallSpec{
		definition: novelprompt.TruthSettlement, step: RunSettling,
		userPayload: map[string]any{
			"chapter_content": bestContent,
			"previous_truth":  rawJSONValue(input.PreviousTruth),
			"chapter_number":  plan.ChapterNumber,
		},
		temperature: 0.2, maxTokens: 4500,
		validate: func() error { return validateSettlement(settlement) },
	}, &settlement)
	if err != nil {
		return ChapterBundle{}, err
	}

	if err := p.emitStage(ctx, input, sink, RunIndexing, "正在生成下一章可用的记忆索引", 92); err != nil {
		return ChapterBundle{}, err
	}
	var memory memoryIndexOutput
	memoryRaw, err := p.callJSON(ctx, input, sink, provider, modelCallSpec{
		definition: novelprompt.MemoryIndex, step: RunIndexing,
		userPayload: map[string]any{
			"chapter_bundle": map[string]any{
				"chapterPlan": plan, "content": bestContent, "audit": bestAudit,
			},
			"previous_memory":  rawJSONValue(input.PreviousMemory),
			"truth_settlement": settlement,
		},
		temperature: 0.2, maxTokens: 3500,
		validate: func() error { return validateMemory(memory) },
	}, &memory)
	if err != nil {
		return ChapterBundle{}, err
	}
	if err := p.checkCancelled(ctx, input); err != nil {
		return ChapterBundle{}, err
	}

	bundle, err := buildChapterBundle(input, selectedVolume, foundationRaw, plan, planRaw, bestContent, bestContentRaw, bestAuditRaw, settlementRaw, memoryRaw, audits, revisions, finalStatus)
	if err != nil {
		return ChapterBundle{}, &DomainError{Code: ErrPersistenceFailed, Message: "failed to assemble chapter bundle", Cause: err}
	}
	if len(localChecks) > 0 {
		var snapshot map[string]json.RawMessage
		if err = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); err != nil {
			return ChapterBundle{}, err
		}
		snapshot["localChecks"] = localChecks
		bundle.RuntimeSnapshot, err = json.Marshal(snapshot)
		if err != nil {
			return ChapterBundle{}, err
		}
	}
	return bundle, nil
}

func modeContract(mode string) map[string]string {
	normalized := strings.TrimSpace(mode)
	contracts := map[string]map[string]string{
		"long_original":   {"id": "long_original", "name": "长篇原创", "rule": "从零建立原创世界、人物弧光与可持续卷章结构；本次输出当前章节，不提前透支后续主线。"},
		"short_original":  {"id": "short_original", "name": "短篇原创", "rule": "围绕单一核心冲突完成自洽短篇；开端、升级、转折和收束必须在有限篇幅内闭环。"},
		"long_imitation":  {"id": "long_imitation", "name": "仿写长篇", "rule": "只学习参考文本的节奏、叙述距离和句式特征；人物、世界、事件与表达必须重新创作，禁止复制原句。"},
		"short_imitation": {"id": "short_imitation", "name": "仿写短篇", "rule": "只迁移参考文本的风格特征，在短篇内完成全新人物、情节和结局，禁止复述原作。"},
		"fanfiction":      {"id": "fanfiction", "name": "同人创作", "rule": "遵守用户给出的既有人物性格、关系和世界规则；新增事件不得让角色无理由失真。"},
		"continuation":    {"id": "continuation", "name": "续写", "rule": "紧接用户提供正文的最后有效事件续写，保持人称、时态、叙述语气和已知事实一致，不重写前文。"},
	}
	if contract, ok := contracts[normalized]; ok {
		return contract
	}
	return contracts["long_original"]
}

func (p *Pipeline) auditChapter(ctx context.Context, input RunContext, sink EventSink, provider batchfactoryv11.DirectorProvider, foundation foundationPlanOutput, plan chapterPlanOutput, content string, round int) (chapterAuditOutput, json.RawMessage, error) {
	var audit chapterAuditOutput
	definition := novelprompt.ChapterAudit
	if input.Agent != nil {
		definition = novelprompt.AgentChapterAudit
	}
	raw, err := p.callJSON(ctx, input, sink, provider, modelCallSpec{
		definition: definition, step: RunAuditing, key: fmt.Sprintf(":%d", round),
		userPayload: withReviewDimensionContext(input, map[string]any{
			"foundation_plan":     foundation,
			"chapter_plan":        plan,
			"chapter_content":     content,
			"pass_score":          p.threshold(input),
			"rules":               input.Rules,
			"notes":               input.Notes,
			"knowledge_materials": input.KnowledgeMaterials,
			"previous_truth":      rawJSONValue(input.PreviousTruth),
			"previous_chapters":   input.PreviousChapters,
			"memory_context":      rawJSONValue(input.PreviousMemory),
		}, foundation),
		temperature: 0.2, maxTokens: 4500,
		validate: func() error { return validateAudit(audit, p.threshold(input)) },
	}, &audit)
	return audit, raw, err
}

func (p *Pipeline) callJSON(ctx context.Context, input RunContext, sink EventSink, provider batchfactoryv11.DirectorProvider, spec modelCallSpec, target any) (json.RawMessage, error) {
	if input.LongImitation != nil && spec.definition.ID != novelprompt.LongImitationRequirements.ID {
		spec.definition.SystemPrompt += novelprompt.LongImitationSystemSupplement
		if payload, ok := spec.userPayload.(map[string]any); ok {
			copy := map[string]any{}
			for k, v := range payload {
				copy[k] = v
			}
			copy["creation_brief"] = imitationModelBrief(input)
			copy["long_imitation_analysis"] = input.ImitationAnalysis
			spec.userPayload = copy
		}
		if input.Project.TargetChapterWords > 8000 && spec.maxTokens >= 12000 {
			spec.maxTokens = input.Project.TargetChapterWords*2 + 4000
		}
	}
	if input.Agent != nil {
		spec.temperature = input.Agent.BaseTemperature
		if spec.definition.ID == novelprompt.ChapterRevision.ID {
			var round int
			fmt.Sscanf(spec.key, ":%d", &round)
			spec.temperature = boundedAgentNumber(spec.temperature+float64(round)*input.Agent.TemperatureStep, 0, 2, 0)
		}
	}
	if input.Agent == nil && spec.definition.ID == novelprompt.ChapterAudit.ID {
		if fields, ok := spec.userPayload.(map[string]any); ok {
			delete(fields, "pass_score")
		}
	}
	userPrompt, err := marshalPromptPayload(withThinkingLevel(spec.userPayload, input.Run.ThinkingLevel))
	if err != nil {
		return nil, &DomainError{Code: ErrInvalidOutput, Message: "failed to build model input", Cause: err}
	}
	request := batchfactoryv11.TextCompletionRequest{
		SystemPrompt: agentRequirementsSystemPrompt(input, spec.definition),
		UserPrompt:   userPrompt,
		Temperature:  spec.temperature,
		MaxTokens:    spec.maxTokens,
	}
	reply, err := p.complete(ctx, input, sink, provider, spec.step, spec.definition.ID+spec.key, request, spec.definition.Version)
	if err != nil {
		return nil, err
	}
	// Apply the same envelope guard before either decode path, including a
	// schema-repair response, so malformed oversize results stay locally cached.
	if input.LongImitation != nil {
		if limitError := checkImitationRawReply(spec.definition.ID, reply); limitError != nil {
			return nil, limitError
		}
	}
	validationErr := decodeAndValidate(reply, target, spec.validate)
	if validationErr == nil {
		return json.RawMessage(reply), nil
	}
	var imitationLimit *imitationOutputLimitError
	if input.LongImitation != nil && errors.As(validationErr, &imitationLimit) {
		return nil, imitationLimitFailure(imitationLimit)
	}

	repairPrompt, marshalErr := marshalPromptPayload(map[string]any{
		"instruction":      novelprompt.StructureRepairInstruction,
		"thinking_level":   strings.TrimSpace(input.Run.ThinkingLevel),
		"original_request": json.RawMessage(userPrompt),
		"invalid_output":   reply,
		"validation_error": validationErr.Error(),
	})
	if marshalErr != nil {
		return nil, invalidOutputError(spec.definition, validationErr)
	}
	repairRequest := batchfactoryv11.TextCompletionRequest{
		SystemPrompt: agentRequirementsSystemPrompt(input, spec.definition) + "\n\n" + novelprompt.StructureRepairSystem,
		UserPrompt:   repairPrompt,
		Temperature:  0.1,
		MaxTokens:    spec.maxTokens,
	}
	repairedReply, err := p.complete(ctx, input, sink, provider, spec.step, spec.definition.ID+spec.key+":repair", repairRequest, spec.definition.Version)
	if err != nil {
		return nil, err
	}
	if input.LongImitation != nil {
		if limitError := checkImitationRawReply(spec.definition.ID, repairedReply); limitError != nil {
			return nil, limitError
		}
	}
	if err := decodeAndValidate(repairedReply, target, spec.validate); err != nil {
		var imitationLimit *imitationOutputLimitError
		if input.LongImitation != nil && errors.As(err, &imitationLimit) {
			return nil, imitationLimitFailure(imitationLimit)
		}
		return nil, invalidOutputError(spec.definition, err)
	}
	return json.RawMessage(repairedReply), nil
}

func (p *Pipeline) complete(ctx context.Context, input RunContext, sink EventSink, provider batchfactoryv11.DirectorProvider, step RunStatus, key string, request batchfactoryv11.TextCompletionRequest, promptVersion int) (string, error) {
	if err := p.checkCancelled(ctx, input); err != nil {
		return "", err
	}
	reply, err := journalCompletePrepared(ctx, key, provider, request, input.Agent != nil && input.Agent.RequireSecondaryConfirm, func() (batchfactoryv11.DirectorProvider, batchfactoryv11.TextCompletionRequest, *ModelCallMetadata, error) {
		if input.Agent == nil {
			id := strings.TrimSpace(input.Run.ModelID)
			if id == "" {
				id = strings.TrimSpace(input.Project.ModelID)
			}
			actual, err := p.resolveProvider(ctx, input.OwnerUsername, id)
			if err != nil {
				return nil, request, nil, mapModelError(err)
			}
			if actual == nil {
				return nil, request, nil, &DomainError{Code: ErrModelUnavailable, Message: "所选文本模型当前不可用"}
			}
			return actual, request, nil, nil
		}
		route := strings.Split(strings.TrimPrefix(key, "novel-creation."), ":")[0]
		routeKey := route
		if route == "style-analysis" || route == "long-imitation-analysis" || route == "long-imitation-requirements" {
			routeKey = "foundation-plan"
		}
		switch route {
		case "short-source-analysis", "short-outline", "short-outline-revision":
			routeKey = "foundation-plan"
		case "short-body":
			routeKey = "chapter-draft"
		case "short-review":
			routeKey = "chapter-audit"
		case "short-body-revision":
			routeKey = "chapter-revision"
		case "short-recommendation", "short-continuity":
			routeKey = "memory-index"
		}
		modelID := input.Agent.ModelRoutes[routeKey]
		if modelID == "" {
			modelID = input.Run.ModelID
		}
		if modelID == "" {
			modelID = input.Project.ModelID
		}
		actual, resolveErr := p.resolveProvider(ctx, input.OwnerUsername, modelID)
		if resolveErr != nil {
			return nil, request, nil, mapModelError(resolveErr)
		}
		if actual == nil {
			return nil, request, nil, &DomainError{Code: ErrModelUnavailable, Message: "所选文本模型当前不可用"}
		}
		meta := ModelCallMetadata{ChosenModelID: modelID, PromptID: "novel-creation." + route, PromptVersion: promptVersion, Temperature: request.Temperature, MaxTokens: request.MaxTokens, TemperatureStatus: "sent", MaxTokensStatus: "sent"}
		if withMeta, ok := actual.(batchfactoryv11.TextModelMetadataProvider); ok {
			meta.Provider = withMeta.TextModelMetadata()
			if meta.Provider.SupportsTemperature != nil && !*meta.Provider.SupportsTemperature {
				request.OmitTemperature = true
				meta.TemperatureStatus = "不支持，未发送"
			}
			if meta.Provider.SupportsMaxTokens != nil && !*meta.Provider.SupportsMaxTokens {
				request.OmitMaxTokens = true
				meta.MaxTokensStatus = "不支持，未发送"
			}
		}
		return actual, request, &meta, nil
	})
	if err != nil {
		return "", mapModelError(err)
	}
	if err := p.checkCancelled(ctx, input); err != nil {
		return "", err
	}
	if strings.TrimSpace(reply) == "" {
		return "", &DomainError{Code: ErrInvalidOutput, Message: "model returned empty output"}
	}
	payload, _ := json.Marshal(map[string]string{"delta": reply})
	if err := p.emit(ctx, input, sink, RunEvent{
		Type: "assistant.delta", Step: string(step), Status: step,
		Message: reply, Payload: payload,
	}); err != nil {
		return "", err
	}
	return reply, nil
}

func (p *Pipeline) validateRunContext(input RunContext, sink EventSink) error {
	if p == nil || p.resolveProvider == nil {
		return &DomainError{Code: ErrModelUnavailable, Message: "novel model resolver is unavailable"}
	}
	if sink == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "run event sink is required"}
	}
	if input.Runtime == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "novel runtime is required"}
	}
	if strings.TrimSpace(input.Run.ID) == "" || strings.TrimSpace(input.Run.ProjectID) == "" || strings.TrimSpace(input.Run.SessionID) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "run id, projectId and sessionId are required"}
	}
	if strings.TrimSpace(input.OwnerUsername) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "ownerUsername is required"}
	}
	if input.Run.ProjectID != input.Project.ID || input.Run.SessionID != input.Session.ID || input.Session.ProjectID != input.Project.ID {
		return &DomainError{Code: ErrInputRequired, Message: "run, session and project bindings do not match"}
	}
	if strings.TrimSpace(input.Run.ModelID) == "" && strings.TrimSpace(input.Project.ModelID) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "modelId is required"}
	}
	if strings.TrimSpace(input.UserMessage) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "user message is required"}
	}
	return nil
}

func (p *Pipeline) checkCancelled(ctx context.Context, input RunContext) error {
	if err := ctx.Err(); err != nil {
		return mapModelError(err)
	}
	cancelled, err := input.Runtime.IsCancelled(ctx, input.Run.ID)
	if err != nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "failed to read cancellation state", Cause: err}
	}
	if cancelled {
		return &DomainError{Code: ErrRunCancelled, Message: "novel run was cancelled", Cause: context.Canceled}
	}
	return nil
}

func (p *Pipeline) emitStage(ctx context.Context, input RunContext, sink EventSink, status RunStatus, message string, progress int) error {
	return p.emit(ctx, input, sink, RunEvent{
		Type: "run.step", Step: string(status), Status: status,
		Message: message, Progress: progress,
	})
}

func (p *Pipeline) emit(ctx context.Context, input RunContext, sink EventSink, event RunEvent) error {
	event.RunID = input.Run.ID
	event.SessionID = input.Run.SessionID
	event.ProjectID = input.Run.ProjectID
	event.CreatedAt = time.Now().UTC()
	persisted, err := sink.Persist(ctx, event)
	if err != nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "failed to persist novel run event", Cause: err}
	}
	if err := sink.Publish(ctx, persisted); err != nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "failed to publish novel run event", Cause: err}
	}
	return nil
}

func outputContractSystemPrompt(definition promptcatalog.Definition) string {
	return strings.TrimSpace(definition.SystemPrompt) + "\n\n输出合同（必须严格遵守）：\n" + strings.TrimSpace(definition.OutputContract)
}

func marshalPromptPayload(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func withThinkingLevel(payload any, thinkingLevel string) any {
	fields, ok := payload.(map[string]any)
	if !ok {
		return map[string]any{"thinking_level": strings.TrimSpace(thinkingLevel), "input": payload}
	}
	copyFields := make(map[string]any, len(fields)+1)
	for key, value := range fields {
		copyFields[key] = value
	}
	copyFields["thinking_level"] = strings.TrimSpace(thinkingLevel)
	return copyFields
}

func decodeAndValidate(raw string, target any, validate func() error) error {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("structured output target must be a non-nil pointer")
	}
	value.Elem().Set(reflect.Zero(value.Elem().Type()))
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("model returned more than one JSON value")
		}
		return err
	}
	if validate != nil {
		return validate()
	}
	return nil
}

func invalidOutputError(definition promptcatalog.Definition, cause error) error {
	return &DomainError{
		Code:    ErrInvalidOutput,
		Message: definition.ID + " returned invalid structured output after one repair",
		Cause:   cause,
	}
}

func mapModelError(err error) error {
	if err == nil {
		return nil
	}
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return err
	}
	if errors.Is(err, context.Canceled) {
		return &DomainError{Code: ErrRunCancelled, Message: "小说创作已取消", Cause: err}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &DomainError{Code: ErrGenerationTimeout, Message: "小说生成超时，请重试", Cause: err}
	}
	if classified := classifyUpstreamModelError(err); classified != nil {
		return classified
	}
	return &DomainError{Code: ErrModelUnavailable, Message: "模型调用失败，请稍后重试", Cause: err}
}

var (
	upstreamHTTPStatus = regexp.MustCompile(`\b(429|502|503|504)\b`)
	upstreamAuthStatus = regexp.MustCompile(`\b(401|403)\b`)
)

// classifyUpstreamModelError turns transport/API wording into actionable,
// user-readable failures instead of one generic "model unavailable" message.
func classifyUpstreamModelError(err error) error {
	text := strings.ToLower(err.Error())
	switch {
	case upstreamAuthStatus.MatchString(text), containsAny(text, "unauthorized", "forbidden", "invalid api key", "incorrect api key", "authentication failed", "鉴权"):
		return &DomainError{Code: ErrModelAuthFailed, Message: "所选文本模型当前不可用，请联系管理员或选择其他授权模型", Cause: err}
	case upstreamHTTPStatus.FindString(text) == "429", containsAny(text, "rate limit", "too many requests", "requests per minute", "限流"):
		return &DomainError{Code: ErrModelRateLimited, Message: "模型服务限流，请稍后重试", Cause: err}
	case upstreamHTTPStatus.MatchString(text),
		containsAny(text, "bad gateway", "service unavailable", "overloaded", "overload", "过载", "上游服务不可用"):
		return &DomainError{Code: ErrModelOverloaded, Message: "模型服务暂时不可用，请稍后重试", Cause: err}
	case containsAny(text, "eof", "connection reset", "connection refused", "broken pipe", "no such host",
		"tls handshake", "server closed", "connection closed", "连接中断", "网络连接"):
		return &DomainError{Code: ErrModelConnectionLost, Message: "模型连接中断，请重试", Cause: err}
	default:
		return nil
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func validateFoundation(value foundationPlanOutput) error {
	for name, text := range map[string]string{
		"title": value.Title, "genre": value.Genre, "world": value.World, "coreConflict": value.CoreConflict,
	} {
		if err := requireText(name, text); err != nil {
			return err
		}
	}
	if len(value.Characters) == 0 || len(value.Volumes) == 0 {
		return errors.New("characters and volumes must be non-empty arrays")
	}
	for index, character := range value.Characters {
		for name, text := range map[string]string{"name": character.Name, "role": character.Role, "goal": character.Goal, "conflict": character.Conflict} {
			if err := requireText(fmt.Sprintf("characters[%d].%s", index, name), text); err != nil {
				return err
			}
		}
	}
	for index, volume := range value.Volumes {
		if volume.Number <= 0 || len(volume.ChapterRange) != 2 || volume.ChapterRange[0] <= 0 || volume.ChapterRange[1] < volume.ChapterRange[0] {
			return fmt.Errorf("volumes[%d] has invalid number or chapterRange", index)
		}
		if err := requireText(fmt.Sprintf("volumes[%d].title", index), volume.Title); err != nil {
			return err
		}
		if err := requireText(fmt.Sprintf("volumes[%d].goal", index), volume.Goal); err != nil {
			return err
		}
	}
	return nil
}

func validateChapterPlan(value chapterPlanOutput, requestedChapter int, volumes []foundationVolume) error {
	if value.ChapterNumber <= 0 {
		return errors.New("chapterNumber must be positive")
	}
	if value.ChapterNumber != requestedChapter {
		return fmt.Errorf("chapterNumber=%d must equal requested chapter %d", value.ChapterNumber, requestedChapter)
	}
	if _, err := outlineVolumeForChapter(volumes, value.ChapterNumber); err != nil {
		return err
	}
	if value.TargetWords < minTargetChapterWords || value.TargetWords > maxTargetChapterWords {
		return fmt.Errorf("targetWords must be between %d and %d", minTargetChapterWords, maxTargetChapterWords)
	}
	for name, text := range map[string]string{
		"title": value.Title, "goal": value.Goal, "conflict": value.Conflict, "turn": value.Turn, "hook": value.Hook,
	} {
		if err := requireText(name, text); err != nil {
			return err
		}
	}
	return nil
}

func outlineVolumeForChapter(volumes []foundationVolume, chapterNumber int) (foundationVolume, error) {
	var matched foundationVolume
	matches := 0
	for _, volume := range volumes {
		if len(volume.ChapterRange) != 2 {
			continue
		}
		if chapterNumber >= volume.ChapterRange[0] && chapterNumber <= volume.ChapterRange[1] {
			matched = volume
			matches++
		}
	}
	if matches == 0 {
		return foundationVolume{}, fmt.Errorf("chapter %d is outside every outline volume", chapterNumber)
	}
	if matches > 1 {
		return foundationVolume{}, fmt.Errorf("chapter %d belongs to more than one outline volume", chapterNumber)
	}
	return matched, nil
}

func validateAudit(value chapterAuditOutput, threshold float64) error {
	if value.Score == nil || *value.Score < 0 || *value.Score > 100 {
		return errors.New("score must be a number between 0 and 100")
	}
	if value.Passed == nil {
		return errors.New("passed must be a boolean")
	}
	if value.Issues == nil {
		return errors.New("issues must be an array")
	}
	seen := make(map[string]struct{}, len(value.Issues))
	for index, issue := range value.Issues {
		if err := requireText(fmt.Sprintf("issues[%d].id", index), issue.ID); err != nil {
			return err
		}
		if _, exists := seen[issue.ID]; exists {
			return fmt.Errorf("duplicate issue id %q", issue.ID)
		}
		seen[issue.ID] = struct{}{}
		switch issue.Severity {
		case "low", "medium", "high", "critical":
		default:
			return fmt.Errorf("issues[%d].severity is invalid", index)
		}
		if err := requireText(fmt.Sprintf("issues[%d].description", index), issue.Description); err != nil {
			return err
		}
		if err := validateOptionalAuditString(fmt.Sprintf("issues[%d].category", index), issue.Category); err != nil {
			return err
		}
		if err := validateOptionalAuditString(fmt.Sprintf("issues[%d].evidence", index), issue.Evidence); err != nil {
			return err
		}
	}
	expectedPassed := *value.Score >= threshold && !hasBlockingAuditIssue(value.Issues)
	if *value.Passed != expectedPassed {
		return fmt.Errorf("passed=%v is inconsistent with score=%g threshold=%g and blocking issues", *value.Passed, *value.Score, threshold)
	}
	return nil
}

func validateOptionalAuditString(name string, raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if strings.TrimSpace(string(raw)) == "null" {
		return fmt.Errorf("%s must be a string when present", name)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s must be a string when present: %w", name, err)
	}
	return nil
}

func validateRevision(value chapterRevisionOutput, currentIssues []auditIssue) error {
	if err := requireText("strategy", value.Strategy); err != nil {
		return err
	}
	if err := requireText("content", value.Content); err != nil {
		return err
	}
	if value.ResolvedIssueIDs == nil {
		return errors.New("resolvedIssueIDs must be an array")
	}
	allowed := make(map[string]struct{}, len(currentIssues))
	for _, issue := range currentIssues {
		allowed[issue.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(value.ResolvedIssueIDs))
	for _, issueID := range value.ResolvedIssueIDs {
		if _, ok := allowed[issueID]; !ok {
			return fmt.Errorf("resolvedIssueIDs contains %q outside the current audit", issueID)
		}
		if _, duplicate := seen[issueID]; duplicate {
			return fmt.Errorf("resolvedIssueIDs contains duplicate %q", issueID)
		}
		seen[issueID] = struct{}{}
	}
	return nil
}

func validateSettlement(value truthSettlementOutput) error {
	if value.Facts == nil || value.CharacterChanges == nil || value.Timeline == nil || value.Foreshadowing == nil {
		return errors.New("settlement fields must all be arrays")
	}
	for index, fact := range value.Facts {
		if err := requireText(fmt.Sprintf("facts[%d].id", index), fact.ID); err != nil {
			return err
		}
		if err := requireText(fmt.Sprintf("facts[%d].statement", index), fact.Statement); err != nil {
			return err
		}
	}
	for index, change := range value.CharacterChanges {
		if err := requireText(fmt.Sprintf("characterChanges[%d].character", index), change.Character); err != nil {
			return err
		}
		if err := requireText(fmt.Sprintf("characterChanges[%d].change", index), change.Change); err != nil {
			return err
		}
	}
	for index, event := range value.Timeline {
		if err := requireText(fmt.Sprintf("timeline[%d].event", index), event.Event); err != nil {
			return err
		}
		if err := requireText(fmt.Sprintf("timeline[%d].position", index), event.Position); err != nil {
			return err
		}
	}
	for index, thread := range value.Foreshadowing {
		if err := requireText(fmt.Sprintf("foreshadowing[%d].thread", index), thread.Thread); err != nil {
			return err
		}
		if thread.Status != "open" && thread.Status != "resolved" {
			return fmt.Errorf("foreshadowing[%d].status is invalid", index)
		}
	}
	return nil
}

func validateMemory(value memoryIndexOutput) error {
	if err := requireText("summary", value.Summary); err != nil {
		return err
	}
	if value.Entities == nil || value.OpenThreads == nil {
		return errors.New("entities and openThreads must be arrays")
	}
	for index, entity := range value.Entities {
		for name, text := range map[string]string{"name": entity.Name, "type": entity.Type, "state": entity.State} {
			if err := requireText(fmt.Sprintf("entities[%d].%s", index, name), text); err != nil {
				return err
			}
		}
	}
	for index, thread := range value.OpenThreads {
		for name, text := range map[string]string{"id": thread.ID, "thread": thread.Thread, "nextExpectation": thread.NextExpectation} {
			if err := requireText(fmt.Sprintf("openThreads[%d].%s", index, name), text); err != nil {
				return err
			}
		}
	}
	return nil
}

func requireText(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must be a non-empty string", name)
	}
	return nil
}

func scoreOf(audit chapterAuditOutput) float64 {
	if audit.Score == nil {
		return 0
	}
	return *audit.Score
}

func hasBlockingAuditIssue(issues []auditIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "high" || issue.Severity == "critical" {
			return true
		}
	}
	return false
}

func auditPasses(audit chapterAuditOutput, threshold float64) bool {
	return audit.Score != nil && audit.Passed != nil && *audit.Score >= threshold && *audit.Passed && !hasBlockingAuditIssue(audit.Issues)
}

func betterAudit(candidate, current chapterAuditOutput, threshold float64) bool {
	if auditPasses(candidate, threshold) && !auditPasses(current, threshold) {
		return true
	}
	return scoreOf(candidate) > scoreOf(current)
}

func rawJSONValue(raw json.RawMessage) any {
	if len(raw) == 0 || !json.Valid(raw) {
		return map[string]any{}
	}
	return json.RawMessage(raw)
}

func newAudit(input RunContext, sequence int, audit chapterAuditOutput, raw json.RawMessage) Audit {
	return Audit{
		ID:        fmt.Sprintf("%s-audit-%d", input.Run.ID, sequence),
		ProjectID: input.Project.ID, RunID: input.Run.ID,
		Score: scoreOf(audit), Snapshot: cloneRawJSON(raw), CreatedAt: time.Now().UTC(),
	}
}

func newRevision(input RunContext, sequence int, revision chapterRevisionOutput, raw json.RawMessage) Revision {
	return Revision{
		ID:        fmt.Sprintf("%s-revision-%d", input.Run.ID, sequence),
		ProjectID: input.Project.ID, RunID: input.Run.ID,
		Strategy: strings.TrimSpace(revision.Strategy), Content: strings.TrimSpace(revision.Content),
		Snapshot: cloneRawJSON(raw), CreatedAt: time.Now().UTC(),
	}
}

// checkpointBundle returns the blueprint as an awaiting-confirmation result
// instead of continuing into chapter writing.
func checkpointBundle(input RunContext, foundationRaw json.RawMessage) (ChapterBundle, error) {
	now := time.Now().UTC()
	snapshot, err := json.Marshal(map[string]any{"plan": json.RawMessage(foundationRaw)})
	if err != nil {
		return ChapterBundle{}, &DomainError{Code: ErrPersistenceFailed, Message: "failed to assemble checkpoint foundation", Cause: err}
	}
	runtimeSnapshot, err := json.Marshal(map[string]any{"foundationPlan": json.RawMessage(foundationRaw)})
	if err != nil {
		return ChapterBundle{}, &DomainError{Code: ErrPersistenceFailed, Message: "failed to assemble checkpoint runtime snapshot", Cause: err}
	}
	return ChapterBundle{
		ProjectID: input.Project.ID,
		RunID:     input.Run.ID,
		Foundation: Foundation{
			ID: input.Project.ID + "-foundation", ProjectID: input.Project.ID,
			Snapshot: snapshot, CreatedAt: now, UpdatedAt: now,
		},
		FinalRunStatus:  RunAwaitingConfirmation,
		RuntimeSnapshot: runtimeSnapshot,
	}, nil
}

func buildChapterBundle(input RunContext, volume foundationVolume, foundationRaw json.RawMessage, plan chapterPlanOutput, planRaw json.RawMessage, content string, contentRaw, auditRaw, settlementRaw, memoryRaw json.RawMessage, audits []Audit, revisions []Revision, finalStatus RunStatus) (ChapterBundle, error) {
	now := time.Now().UTC()
	volumeID := fmt.Sprintf("%s-volume-%d", input.Project.ID, volume.Number)
	chapterID := fmt.Sprintf("%s-chapter-%d", input.Project.ID, plan.ChapterNumber)
	foundationSnapshot, err := json.Marshal(map[string]any{
		"plan": json.RawMessage(foundationRaw), "truthSettlement": json.RawMessage(settlementRaw),
	})
	if err != nil {
		return ChapterBundle{}, err
	}
	versionSnapshot, err := json.Marshal(map[string]any{
		"truthSettlement": json.RawMessage(settlementRaw), "memoryIndex": json.RawMessage(memoryRaw),
		"chapterPlan": json.RawMessage(planRaw), "modelOutput": json.RawMessage(contentRaw), "finalAudit": json.RawMessage(auditRaw),
	})
	if err != nil {
		return ChapterBundle{}, err
	}
	runtimeSnapshot, err := json.Marshal(map[string]any{
		"foundationPlan": json.RawMessage(foundationRaw), "chapterPlan": json.RawMessage(planRaw),
		"finalAudit": json.RawMessage(auditRaw), "truthSettlement": json.RawMessage(settlementRaw),
		"memoryIndex": json.RawMessage(memoryRaw),
	})
	if err != nil {
		return ChapterBundle{}, err
	}
	auditStatus := "passed"
	if finalStatus == RunSucceededWithWarnings {
		auditStatus = "warning"
	}
	return ChapterBundle{
		ProjectID: input.Project.ID,
		RunID:     input.Run.ID,
		Volume:    Volume{ID: volumeID, ProjectID: input.Project.ID, Number: volume.Number, Title: volume.Title, CreatedAt: now, UpdatedAt: now},
		Chapter: Chapter{
			ID: chapterID, ProjectID: input.Project.ID, VolumeID: volumeID,
			Number: plan.ChapterNumber, Title: plan.Title, Content: strings.TrimSpace(content),
			WordCount: countVisibleRunes(content), AuditStatus: auditStatus, Version: 1,
			CreatedAt: now, UpdatedAt: now,
		},
		ChapterVersion: ChapterVersion{
			ID: chapterID + "-version-" + input.Run.ID, ChapterID: chapterID, RunID: input.Run.ID,
			Version: 1, Content: strings.TrimSpace(content), Snapshot: versionSnapshot, CreatedAt: now,
		},
		Foundation: Foundation{ID: input.Project.ID + "-foundation", ProjectID: input.Project.ID, Snapshot: foundationSnapshot, CreatedAt: now, UpdatedAt: now},
		Memories: []Memory{{
			ID: input.Run.ID + "-memory", ProjectID: input.Project.ID, ChapterID: chapterID,
			Snapshot: cloneRawJSON(memoryRaw), CreatedAt: now, UpdatedAt: now,
		}},
		Audits: audits, Revisions: revisions, FinalRunStatus: finalStatus, RuntimeSnapshot: runtimeSnapshot,
	}, nil
}

func countVisibleRunes(value string) int {
	count := 0
	for _, current := range value {
		if !unicode.IsSpace(current) {
			count++
		}
	}
	return count
}

func (p *Pipeline) threshold(input RunContext) float64 {
	if input.Agent != nil {
		return input.Agent.PassScore
	}
	return p.auditThreshold
}
