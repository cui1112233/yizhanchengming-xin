package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const shortStoryKind = "short_story"

type ShortStoryInput struct {
	IdempotencyKey    string                   `json:"idempotencyKey"`
	Mode              string                   `json:"mode"`
	ModelID           string                   `json:"modelId"`
	ThinkingLevel     string                   `json:"thinkingLevel"`
	Title             string                   `json:"title"`
	ReferenceText     string                   `json:"referenceText"`
	Requirements      string                   `json:"requirements"`
	TargetWordCount   int                      `json:"targetWordCount"`
	TargetChapters    int                      `json:"targetChapters"`
	ReaderPreferences string                   `json:"readerPreferences"`
	SupplementaryInfo string                   `json:"supplementaryInfo"`
	InspirationTags   []string                 `json:"inspirationTags"`
	KnowledgeBookIDs  []string                 `json:"knowledgeBookIds"`
	MemoReferences    []ImitationMemoReference `json:"memoReferences"`
	LinkedSaveID      string                   `json:"linkedSaveId"`
	PauseAfterOutline bool                     `json:"pauseAfterOutline"`
}
type ShortOutlineChapter struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
type ShortOutline struct {
	Title    string                `json:"title"`
	Summary  string                `json:"summary"`
	Chapters []ShortOutlineChapter `json:"chapters"`
}
type ShortChapter struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
type ShortArtifact struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
type ShortStory struct {
	ID             string            `json:"storyId"`
	ProjectID      string            `json:"projectId"`
	SessionID      string            `json:"sessionId"`
	RunID          string            `json:"runId"`
	Version        int               `json:"version"`
	Status         string            `json:"status"`
	Title          string            `json:"title"`
	Input          ShortStoryInput   `json:"input"`
	Outline        *ShortOutline     `json:"outline"`
	Chapters       []ShortChapter    `json:"chapters"`
	Artifacts      []ShortArtifact   `json:"artifacts"`
	Recommendation string            `json:"recommendation"`
	Styles         []StyleProfile    `json:"styles"`
	Notes          []ProjectNote     `json:"notes"`
	LinkedContext  string            `json:"linkedContext,omitempty"`
	OutlineAudit   json.RawMessage   `json:"outlineAudit,omitempty"`
	ChapterAudits  []json.RawMessage `json:"chapterAudits"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}
type ShortStoryResult struct {
	Project Project    `json:"project"`
	Session Session    `json:"session"`
	Run     Run        `json:"run"`
	Story   ShortStory `json:"story"`
}

func normalizeShortStory(in ShortStoryInput) (ShortStoryInput, error) {
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if e := studioKey(in.IdempotencyKey); e != nil {
		return in, e
	}
	in.Mode = strings.TrimSpace(in.Mode)
	if in.Mode != "short_imitation" && in.Mode != "short_original" {
		return in, inputError("请选择短篇创作模式")
	}
	in.ModelID = strings.TrimSpace(in.ModelID)
	if in.ModelID == "" || len(in.ModelID) > 96 {
		return in, inputError("请选择文本模型")
	}
	var e error
	in.ThinkingLevel, e = normalizeStudioThinking(in.ThinkingLevel)
	if e != nil {
		return in, e
	}
	in.Title = jsTrim(in.Title)
	in.ReferenceText = jsTrim(in.ReferenceText)
	in.Requirements = jsTrim(in.Requirements)
	if in.Mode == "short_imitation" && jsNonspaceLength(in.ReferenceText) < 100 {
		return in, inputError("参考原文至少需要100字（不计空白）")
	}
	if in.Mode == "short_original" && in.Requirements == "" {
		return in, inputError("请输入短篇创作要求")
	}
	if in.TargetWordCount < 1000 {
		in.TargetWordCount = 13000
	}
	if in.TargetChapters < 1 {
		in.TargetChapters = 1
	}
	if in.TargetWordCount > 500000 || in.TargetChapters > 50 {
		return in, inputError("总字数最多500000字，篇章最多50章")
	}
	if jsLength(in.ReferenceText) > LongImitationMaxUTF16 || jsLength(in.Requirements) > 50000 || jsLength(in.SupplementaryInfo) > 50000 || jsLength(in.ReaderPreferences) > 12000 || jsLength(in.Title) > 255 || len(in.InspirationTags) > 32 || len(in.KnowledgeBookIDs) > 8 || len(in.MemoReferences) > 8 {
		return in, inputError("输入或引用数量超出上限")
	}
	for _, v := range in.InspirationTags {
		if jsLength(v) > 80 {
			return in, inputError("灵感标签过长")
		}
	}
	return in, nil
}
func (s *Service) StartShortStory(ctx context.Context, owner string, input ShortStoryInput) (out ShortStoryResult, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	in, err := normalizeShortStory(input)
	if err != nil {
		return out, err
	}
	digest := fingerprint(in)
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		receipt, e := tx.ReserveSubmission(ctx, owner, shortStoryKind, in.IdempotencyKey, digest)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != digest {
			return conflict()
		}
		if len(receipt.Result) > 0 {
			return json.Unmarshal(receipt.Result, &out)
		}
		story := ShortStory{ID: novelID("story"), Version: 1, Status: "queued", Title: in.Title, Input: in, Chapters: []ShortChapter{}, Artifacts: []ShortArtifact{}, Styles: []StyleProfile{}, Notes: []ProjectNote{}, CreatedAt: time.Now().UTC()}
		seen := map[string]bool{}
		for _, id := range in.KnowledgeBookIDs {
			if seen[id] || id == "" {
				return inputError("文风选择无效或重复")
			}
			seen[id] = true
			p, e := tx.GetStyleProfile(ctx, owner, id)
			if e != nil {
				return e
			}
			story.Styles = append(story.Styles, p)
		}
		for _, ref := range in.MemoReferences {
			key := memoryKey(ref.ProjectID, ref.NoteID)
			if seen[key] {
				return inputError("备忘录重复")
			}
			seen[key] = true
			notes, e := tx.ListNotes(ctx, owner, ref.ProjectID)
			if e != nil {
				return e
			}
			found := false
			for _, n := range notes {
				if n.ID == ref.NoteID {
					story.Notes = append(story.Notes, n)
					found = true
					break
				}
			}
			if !found {
				return inputError("备忘录不存在或不属于当前账号")
			}
		}
		if in.LinkedSaveID != "" {
			p, e := tx.GetProject(ctx, owner, in.LinkedSaveID)
			if e != nil {
				return e
			}
			story.LinkedContext = strings.Join([]string{p.Title, p.World, p.Protagonist, p.CoreConflict}, "\n")
		}
		title := in.Title
		if title == "" {
			title = "待命名短篇"
		}
		p, e := tx.CreateProject(ctx, owner, CreateProjectInput{Title: title, Genre: "短篇", ModelID: in.ModelID, TargetChapterWords: 2000})
		if e != nil {
			return e
		}
		session, e := tx.CreateSession(ctx, owner, CreateSessionInput{ProjectID: p.ID, Mode: in.Mode, Title: title})
		if e != nil {
			return e
		}
		story.ProjectID, story.SessionID = p.ID, session.ID
		settings, e := tx.GetProjectSettings(ctx, owner, p.ID)
		if e != nil {
			return e
		}
		settings.Agent = withDefaultReviewDimensions(NormalizeAgentSettings(settings.Agent, p.TargetChapterWords))
		// Content checkpoints are distinct from approval of individual paid requests.
		raw, _ := json.Marshal(map[string]any{"short_story": story, "project_settings": settings, "agent_contract_version": 1, "mode": in.Mode})
		run, e := tx.CreateRun(ctx, owner, Run{ProjectID: p.ID, SessionID: session.ID, ModelID: in.ModelID, ThinkingLevel: in.ThinkingLevel, IdempotencyKey: "short:" + fingerprint([]string{owner, in.IdempotencyKey}), Status: RunQueued, PromptSnapshot: raw})
		if e != nil {
			return e
		}
		story.RunID = run.ID
		story.UpdatedAt = story.CreatedAt
		if e = putContent(ctx, tx, owner, p.ID, shortStoryKind, "story", "queued", time.Time{}, story); e != nil {
			return e
		}
		if e = tx.PutExecution(ctx, Execution{Owner: owner, ProjectID: p.ID, RunID: run.ID, State: "ready", Fingerprint: digest, Calls: map[string]CallRecord{}}); e != nil {
			return e
		}
		if _, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: run.ID, ProjectID: p.ID, SessionID: session.ID, Type: "run.queued", Status: RunQueued, Message: "短篇任务已排队"}); e != nil {
			return e
		}
		out = ShortStoryResult{p, session, run, story}
		result, _ := json.Marshal(out)
		return tx.CompleteSubmission(ctx, owner, shortStoryKind, in.IdempotencyKey, submissionReceipt{Fingerprint: digest, ProjectID: p.ID, Result: result})
	})
	if err == nil {
		s.Notify(ctx, owner, out.Run.ID)
	}
	return
}
func (s *Service) GetShortStoryRequest(ctx context.Context, owner, key string) (out ShortStoryResult, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	if err = studioKey(key); err != nil {
		return
	}
	r, ok, err := s.store.GetSubmissionReceipt(ctx, owner, shortStoryKind, key)
	if err != nil {
		return out, err
	}
	if !ok {
		return out, &DomainError{Code: ErrSessionNotFound, Message: "未找到原短篇请求"}
	}
	err = json.Unmarshal(r.Result, &out)
	if err != nil {
		return
	}
	out.Story, err = s.GetShortStory(ctx, owner, out.Project.ID)
	if err == nil {
		out.Run, err = s.store.GetRun(ctx, owner, out.Story.RunID)
	}
	return
}
func (s *Service) GetShortStory(ctx context.Context, owner, project string) (out ShortStory, err error) {
	ok, e := readContent(ctx, s.store, owner, project, shortStoryKind, "story", &out)
	if e != nil {
		return out, e
	}
	if !ok {
		return out, &DomainError{Code: ErrSessionNotFound, Message: "此作品没有独立短篇资料"}
	}
	if out.RunID != "" {
		run, e := s.store.GetRun(ctx, owner, out.RunID)
		if e != nil {
			return out, e
		}
		if run.Status == RunFailed || run.Status == RunCancelled || run.Status == RunPaused {
			out.Status = string(run.Status)
		}
	}
	return out, nil
}
func validateShortOutline(v ShortOutline, count int) error {
	if strings.TrimSpace(v.Title) == "" || strings.TrimSpace(v.Summary) == "" || len(v.Chapters) != count || jsLength(v.Title) > 255 || jsLength(v.Summary) > 16000 {
		return inputError("短篇大纲须包含标题、摘要及准确的目标章节数")
	}
	for i, c := range v.Chapters {
		if c.Number != i+1 || strings.TrimSpace(c.Title) == "" || strings.TrimSpace(c.Summary) == "" || jsLength(c.Summary) > 8000 {
			return inputError("大纲章节顺序、标题或梗概无效")
		}
	}
	return nil
}
func (s *Service) EditShortStoryOutline(ctx context.Context, owner, project string, version int, outline ShortOutline) (out ShortStory, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		child := *s
		child.store = tx
		v, e := child.GetShortStory(ctx, owner, project)
		if e != nil {
			return e
		}
		if v.Version != version {
			return conflict()
		}
		if v.Status != "awaiting_confirmation" {
			return inputError("当前大纲不在可编辑的确认阶段")
		}
		if _, e = writableSession(ctx, tx, owner, project, v.SessionID); e != nil {
			return e
		}
		if e = validateShortOutline(outline, v.Input.TargetChapters); e != nil {
			return e
		}
		v.Outline = &outline
		v.Title = outline.Title
		v.Version++
		v.UpdatedAt = time.Now().UTC()
		out = v
		return putContent(ctx, tx, owner, project, shortStoryKind, "story", v.Status, time.Time{}, v)
	})
	return
}
func (s *Service) ConfirmShortStory(ctx context.Context, owner, id string, version int) (out Run, err error) {
	source, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return
	}
	return s.queueAtomic(ctx, owner, source.ProjectID, fingerprint(version), func(child *Service) (Run, error) {
		key := "confirm-short:" + id
		if run, ok, e := child.store.FindRunByIdempotencyKey(ctx, owner, key); e != nil {
			return Run{}, e
		} else if ok {
			return run, nil
		}
		v, e := child.GetShortStory(ctx, owner, source.ProjectID)
		if e != nil {
			return Run{}, e
		}
		if v.Version != version {
			return Run{}, conflict()
		}
		if v.Status != "awaiting_confirmation" || source.Status != RunAwaitingConfirmation || v.Outline == nil {
			return Run{}, inputError("短篇尚未等待内容确认")
		}
		if e = validateShortOutline(*v.Outline, v.Input.TargetChapters); e != nil {
			return Run{}, e
		}
		if _, e = writableSession(ctx, child.store, owner, source.ProjectID, source.SessionID); e != nil {
			return Run{}, e
		}
		var snap map[string]any
		if e = json.Unmarshal(source.PromptSnapshot, &snap); e != nil {
			return Run{}, e
		}
		snap["short_story"] = v
		snap["short_outline_confirmed"] = true
		raw, _ := json.Marshal(snap)
		run, e := child.store.CreateRun(ctx, owner, Run{ProjectID: source.ProjectID, SessionID: source.SessionID, ModelID: source.ModelID, ThinkingLevel: source.ThinkingLevel, IdempotencyKey: key, Status: RunQueued, PromptSnapshot: raw})
		if e != nil {
			return Run{}, e
		}
		v.RunID = run.ID
		v.Status = "queued"
		v.Version++
		if e = putContent(ctx, child.store, owner, v.ProjectID, shortStoryKind, "story", v.Status, time.Time{}, v); e != nil {
			return Run{}, e
		}
		if _, e = child.store.UpdateRun(ctx, owner, RunUpdate{ID: id, Status: RunSucceeded}); e != nil {
			return Run{}, e
		}
		return run, nil
	})
}
