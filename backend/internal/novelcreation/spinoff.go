package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"qiantie/backend/internal/batchfactoryv11"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

const spinoffKind = "spinoff"

type SpinoffInput struct {
	IdempotencyKey   string   `json:"idempotencyKey"`
	ParentBookID     string   `json:"parentBookId"`
	Title            string   `json:"title"`
	SpinoffType      string   `json:"spinoffType"`
	StoryIdea        string   `json:"storyIdea"`
	ReferenceText    string   `json:"referenceText"`
	KnowledgeBookIDs []string `json:"knowledgeBookIds"`
	ModelID          string   `json:"modelId"`
	ThinkingLevel    string   `json:"thinkingLevel"`
}
type SpinoffCreation struct {
	Input            SpinoffInput             `json:"input"`
	Parent           Project                  `json:"parent"`
	ParentFoundation json.RawMessage          `json:"parentFoundation"`
	ParentChapters   []Chapter                `json:"parentChapters"`
	ParentCanon      string                   `json:"parentCanon"`
	Foundation       json.RawMessage          `json:"foundation"`
	Styles           []importedStyleReference `json:"styles"`
	CreatedAt        time.Time                `json:"createdAt"`
}
type SpinoffResult struct {
	Book     Project         `json:"book"`
	Session  Session         `json:"session"`
	Run      Run             `json:"run"`
	Creation SpinoffCreation `json:"creation"`
}

func (s *Service) StartSpinoff(ctx context.Context, owner string, in SpinoffInput) (out SpinoffResult, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	// Freeze caller-owned slices before normalization or concurrent receipt checks.
	in.KnowledgeBookIDs = append([]string(nil), in.KnowledgeBookIDs...)
	in.Title = strings.TrimSpace(in.Title)
	in.ParentBookID = strings.TrimSpace(in.ParentBookID)
	in.SpinoffType = strings.TrimSpace(in.SpinoffType)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	in.ModelID = strings.TrimSpace(in.ModelID)
	in.StoryIdea = strings.TrimSpace(in.StoryIdea)
	in.ReferenceText = jsTrim(in.ReferenceText)
	if err = studioKey(in.IdempotencyKey); err != nil {
		return
	}
	if in.Title == "" || utf8.RuneCountInString(in.Title) > 255 {
		return out, inputError("请填写书名（最多255字）")
	}
	if in.ParentBookID == "" {
		return out, inputError("请选择正传书籍")
	}
	if in.ModelID == "" || len(in.ModelID) > 96 {
		return out, inputError("请选择本系统可用的文本模型")
	}
	if in.SpinoffType == "" {
		in.SpinoffType = "sequel"
	}
	if in.SpinoffType != "sequel" && in.SpinoffType != "prequel" && in.SpinoffType != "parallel" && in.SpinoffType != "if" {
		return out, inputError("请选择有效的番外类型")
	}
	if jsLength(in.StoryIdea) > 12000 || jsLength(in.ReferenceText) > 50000 || len(in.KnowledgeBookIDs) > 100 {
		return out, inputError("番外方向最多12000字元，参考原文最多50000字元，请减少资料后重试")
	}
	in.ThinkingLevel, err = normalizeStudioThinking(in.ThinkingLevel)
	if err != nil {
		return
	}
	digest := fingerprint(in)
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		receipt, e := tx.ReserveSubmission(ctx, owner, spinoffKind, in.IdempotencyKey, digest)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != digest {
			return conflict()
		}
		if len(receipt.Result) > 0 {
			return persistenceError(json.Unmarshal(receipt.Result, &out))
		}
		parent, e := tx.GetProject(ctx, owner, in.ParentBookID)
		if e != nil {
			return e
		}
		if parent.Status == "archived" || parent.Status == "purging" {
			return inputError("正传已归档或正在清理，请重新选择")
		}
		chapters, e := tx.ListChapters(ctx, owner, parent.ID)
		if e != nil {
			return e
		}
		foundation, found, e := tx.GetFoundation(ctx, owner, parent.ID)
		if e != nil {
			return e
		}
		if len(chapters) == 0 && !found && parent.World == "" && parent.Protagonist == "" {
			return inputError("正传还没有已保存的章节或设定，请先完成正传资料")
		}
		// Bound the stored input without silently discarding events from the parent.
		size := len(parent.World) + len(parent.Protagonist) + len(foundation.Snapshot)
		for _, c := range chapters {
			size += len(c.Content)
		}
		if size > 1000000 {
			return inputError("正传资料超过本次可处理范围，请先整理正传设定后重试")
		}
		creation := SpinoffCreation{Input: in, Parent: parent, ParentChapters: chapters, ParentFoundation: foundation.Snapshot, CreatedAt: time.Now().UTC()}
		settingsStyle := []string{}
		seen := map[string]bool{}
		for _, id := range in.KnowledgeBookIDs {
			if seen[id] || strings.TrimSpace(id) == "" {
				return inputError("文风选择重复或无效")
			}
			seen[id] = true
			style, e := tx.GetStyleProfile(ctx, owner, id)
			if e != nil {
				return e
			}
			if strings.TrimSpace(style.Content) == "" {
				return inputError("所选文风内容为空")
			}
			creation.Styles = append(creation.Styles, importedStyleReference{Profile: style, Digest: fingerprint(style)})
			settingsStyle = append(settingsStyle, style.Content)
		}
		book, e := tx.CreateProject(ctx, owner, CreateProjectInput{Title: in.Title, Genre: "other", ModelID: in.ModelID, TargetChapterWords: 2000})
		if e != nil {
			return e
		}
		if e = inheritAgentDefaults(ctx, tx, owner, book); e != nil {
			return e
		}
		settings, e := tx.GetProjectSettings(ctx, owner, book.ID)
		if e != nil {
			return e
		}
		settings.StyleProfile = strings.Join(settingsStyle, "\n\n")
		if _, e = tx.UpdateProjectSettings(ctx, owner, book.ID, settings); e != nil {
			return e
		}
		if e = putContent(ctx, tx, owner, book.ID, spinoffKind, "creation", "frozen", time.Time{}, creation); e != nil {
			return e
		}
		session, e := tx.CreateSession(ctx, owner, CreateSessionInput{ProjectID: book.ID, Mode: "fanfiction", Title: in.Title})
		if e != nil {
			return e
		}
		prompts, e := s.snapshotSystemPrompts()
		if e != nil {
			return e
		}
		snap, _ := json.Marshal(map[string]any{"spinoff_init": creation, "system_prompts": prompts, "project_settings": settings})
		run, e := tx.CreateRun(ctx, owner, Run{ProjectID: book.ID, SessionID: session.ID, IdempotencyKey: "spinoff:" + fingerprint([]string{owner, in.IdempotencyKey}), Status: RunQueued, ModelID: in.ModelID, ThinkingLevel: in.ThinkingLevel, PromptSnapshot: snap})
		if e != nil {
			return e
		}
		if _, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: run.ID, ProjectID: book.ID, SessionID: session.ID, Type: "run.queued", Status: RunQueued, Message: "番外初始化已排队：整理正传正典与番外设定"}); e != nil {
			return e
		}
		if e = tx.PutExecution(ctx, Execution{Owner: owner, ProjectID: book.ID, RunID: run.ID, State: "ready", Fingerprint: digest, Calls: map[string]CallRecord{}}); e != nil {
			return e
		}
		out = SpinoffResult{book, session, run, creation}
		raw, e := json.Marshal(out)
		if e != nil {
			return e
		}
		return tx.CompleteSubmission(ctx, owner, spinoffKind, in.IdempotencyKey, submissionReceipt{Fingerprint: digest, ProjectID: book.ID, Result: raw})
	})
	if err == nil {
		s.Notify(ctx, owner, out.Run.ID)
	}
	return
}
func loadSpinoff(ctx context.Context, st Store, owner, project string) (*SpinoffCreation, error) {
	var c SpinoffCreation
	ok, e := readContent(ctx, st, owner, project, spinoffKind, "creation", &c)
	if e != nil || !ok {
		return nil, e
	}
	return &c, nil
}
func (s *Service) GetSpinoff(ctx context.Context, owner, project string) (*SpinoffCreation, error) {
	if _, e := s.store.GetProject(ctx, owner, project); e != nil {
		return nil, e
	}
	return loadSpinoff(ctx, s.store, owner, project)
}
func (s *Service) runSpinoff(ctx context.Context, owner string, run Run, c SpinoffCreation, prompts *SystemPromptSnapshot, settings ProjectSettings) (Run, error) {
	p, ok := s.pipeline.(*Pipeline)
	if !ok {
		return s.finishWithError(ctx, owner, run, ErrModelUnavailable, "番外生成服务未配置", nil)
	}
	book, e := s.store.GetProject(ctx, owner, run.ProjectID)
	if e != nil {
		return run, e
	}
	session, e := s.store.GetSession(ctx, owner, run.SessionID)
	if e != nil {
		return run, e
	}
	agent := NormalizeAgentSettings(settings.Agent, book.TargetChapterWords)
	in := RunContext{OwnerUsername: owner, Run: run, Project: book, Session: session, Runtime: s.runtime, SystemPrompts: prompts, Agent: &agent}
	sink := serviceEventSink{service: s, owner: owner}
	fail := func(e error) (Run, error) {
		if isDurableWriteFailure(e) || IsCode(e, ErrPersistenceFailed) {
			return s.recoverDurableWriteFailure(ctx, owner, run, e)
		}
		if IsCode(e, ErrSecondaryConfirmation) {
			return s.finishModelCallConfirmation(ctx, owner, run)
		}
		code, msg := domainFailure(e)
		if errors.Is(e, context.Canceled) {
			code = ErrRunCancelled
			msg = "番外初始化已停止"
		}
		return s.finishWithError(ctx, owner, run, code, msg, e)
	}
	if e = p.emitStage(ctx, in, sink, RunPlanning, "正在整理正传正典", 15); e != nil {
		return fail(e)
	}
	body, e := frozenSystemText(in, novelprompt.SpinoffCanonID, novelprompt.SpinoffCanonSystem)
	if e != nil {
		return fail(e)
	}
	materials, _ := json.Marshal(map[string]any{"parent": c.Parent, "foundation": rawJSONValue(c.ParentFoundation), "chapters": c.ParentChapters})
	// Recovered generator caps parent material at 30000 characters.
	materialText := string(materials)
	runes := []rune(materialText)
	if len(runes) > 30000 {
		materialText = string(runes[:30000])
	}
	user, _ := marshalPromptPayload(map[string]any{"parent_title": c.Parent.Title, "spinoff_type": c.Input.SpinoffType, "materials": materialText, "materialTruncated": len(runes) > 30000})
	canon, e := p.complete(ctx, in, sink, nil, RunPlanning, novelprompt.SpinoffCanonID, batchfactoryv11.TextCompletionRequest{SystemPrompt: body, UserPrompt: user, Temperature: .35, MaxTokens: 9000}, 1)
	if e != nil {
		return fail(e)
	}
	if strings.TrimSpace(canon) == "" {
		return fail(&DomainError{Code: ErrInvalidOutput, Message: "正传正典返回为空，未生成番外设定"})
	}
	if e = p.emitStage(ctx, in, sink, RunPlanning, "正在生成番外设定", 60); e != nil {
		return fail(e)
	}
	var f foundationPlanOutput
	raw, e := p.callJSON(ctx, in, sink, nil, modelCallSpec{definition: novelprompt.SpinoffFoundation, step: RunPlanning, userPayload: map[string]any{"parent_canon": canon, "spinoff_type": c.Input.SpinoffType, "story_idea": c.Input.StoryIdea, "title": c.Input.Title, "reference_text": c.Input.ReferenceText, "style_profile": settings.StyleProfile, "rules": settings.Rules}, temperature: .4, maxTokens: 6000, validate: func() error { return validateFoundation(f) }}, &f)
	if e != nil {
		return fail(e)
	}
	c.ParentCanon = canon
	c.Foundation = raw
	bundle, e := checkpointBundle(in, raw)
	if e != nil {
		return fail(e)
	}
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, &DomainError{Code: ErrPersistenceFailed, Message: "番外保存缺少任务写入凭据"}
	}
	now := time.Now().UTC()
	out, event, e := g.finishRun(ctx, owner, RunUpdate{ID: run.ID, Status: RunSucceeded, RuntimeSnapshot: bundle.RuntimeSnapshot, FinishedAt: &now}, &bundle.Foundation, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.completed", Status: RunSucceeded, Message: "番外设定已生成，可以开始正文创作", Progress: 100}, func(tx Store) error {
		return putContent(ctx, tx, owner, run.ProjectID, spinoffKind, "creation", "ready", time.Time{}, c)
	})
	if e != nil {
		return s.recoverDurableWriteFailure(ctx, owner, run, e)
	}
	if s.runtime != nil {
		_ = s.runtime.Publish(ctx, event)
	}
	return out, nil
}
