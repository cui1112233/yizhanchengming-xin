package novelcreation

import (
	"context"
	"encoding/json"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"time"
	"unicode/utf8"
)

const longImitationKind = "long_imitation"
const LongImitationMaxUTF16 = 2000000

type ImitationMemoReference struct {
	ProjectID string `json:"projectId"`
	NoteID    string `json:"noteId"`
}
type ImitationReferenceSelection struct {
	Files []ImitationSelectedFile `json:"files"`
}
type ImitationSelectedFile struct {
	Filename   string   `json:"filename"`
	ChapterIDs []string `json:"chapterIds"`
}
type LongImitationInput struct {
	IdempotencyKey     string                       `json:"idempotencyKey"`
	ModelID            string                       `json:"modelId"`
	ThinkingLevel      string                       `json:"thinkingLevel"`
	Title              string                       `json:"title"`
	ReferenceText      string                       `json:"referenceText"`
	StoryIdea          string                       `json:"storyIdea"`
	SelectedDirection  string                       `json:"selectedDirection"`
	TargetWordCount    int                          `json:"targetWordCount"`
	TargetChapters     int                          `json:"targetChapters"`
	ReaderPreferences  string                       `json:"readerPreferences"`
	SupplementaryInfo  string                       `json:"supplementaryInfo"`
	KnowledgeBookIDs   []string                     `json:"knowledgeBookIds"`
	MemoReferences     []ImitationMemoReference     `json:"memoReferences"`
	ReferenceSelection *ImitationReferenceSelection `json:"referenceSelection,omitempty"`
}
type LongImitationCreation struct {
	Version                int                        `json:"version"`
	Input                  LongImitationInput         `json:"input"`
	EffectiveReferenceText string                     `json:"effectiveReferenceText"`
	ReferenceChunks        []json.RawMessage          `json:"referenceChunks"`
	ChunkingMethod         string                     `json:"chunkingMethod"`
	Statistics             ImitationStatistics        `json:"statistics"`
	StyleDimensions        []ImitationDimension       `json:"styleDimensions"`
	Styles                 []importedStyleReference   `json:"styles"`
	Memos                  []ProjectNote              `json:"memos"`
	AutoTitle              bool                       `json:"autoTitle"`
	ProvisionalTitle       string                     `json:"provisionalTitle,omitempty"`
	CreatedAt              time.Time                  `json:"createdAt"`
	Analysis               *LongImitationAnalysis     `json:"analysis"`
	Requirements           *LongImitationRequirements `json:"requirements"`
}
type LongImitationResult struct {
	Project  Project               `json:"project"`
	Session  Session               `json:"session"`
	Run      Run                   `json:"run"`
	Creation LongImitationCreation `json:"creation"`
}

func normalizeLongImitation(in LongImitationInput) (LongImitationInput, error) {
	if jsLength(in.ReferenceText) > LongImitationMaxUTF16 {
		return in, inputError("参考原文不能超过200万个UTF16字元（含空白）")
	}
	raw, _ := json.Marshal(in)
	_ = json.Unmarshal(raw, &in)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	in.ModelID = strings.TrimSpace(in.ModelID)
	in.Title = jsTrim(in.Title)
	in.ReferenceText = jsTrim(in.ReferenceText)
	in.StoryIdea = jsTrim(in.StoryIdea)
	in.SelectedDirection = jsTrim(in.SelectedDirection)
	in.ReaderPreferences = jsTrim(in.ReaderPreferences)
	in.SupplementaryInfo = jsTrim(in.SupplementaryInfo)
	if e := studioKey(in.IdempotencyKey); e != nil {
		return in, e
	}
	if in.ModelID == "" || len(in.ModelID) > 96 {
		return in, inputError("请选择文本模型")
	}
	var e error
	in.ThinkingLevel, e = normalizeStudioThinking(in.ThinkingLevel)
	if e != nil {
		return in, e
	}
	if jsNonspaceLength(in.ReferenceText) < 100 {
		return in, inputError("参考原文至少需要100字（去空白UTF16）")
	}
	if jsLength(in.ReferenceText) > LongImitationMaxUTF16 {
		return in, inputError("参考原文不能超过200万个UTF16字元，请缩小选择范围")
	}
	if in.TargetWordCount == 0 {
		in.TargetWordCount = 5000
	}
	if in.TargetChapters == 0 {
		in.TargetChapters = 100
	}
	if in.TargetWordCount < 1000 || in.TargetWordCount > 50000 || in.TargetChapters < 1 || in.TargetChapters > 500 {
		return in, inputError("每章目标须为1000至50000字，目标章数须为1至500章")
	}
	if utf8.RuneCountInString(in.Title) > 255 || jsLength(in.StoryIdea) > 50000 || jsLength(in.ReaderPreferences) > 12000 || jsLength(in.SupplementaryInfo) > 50000 || jsLength(in.SelectedDirection) > 1000 || len(in.KnowledgeBookIDs) > 8 || len(in.MemoReferences) > 8 {
		return in, inputError("创作字段或引用数量超出上限")
	}
	if in.KnowledgeBookIDs == nil {
		in.KnowledgeBookIDs = []string{}
	}
	if in.MemoReferences == nil {
		in.MemoReferences = []ImitationMemoReference{}
	}
	if in.ReferenceSelection != nil {
		if len(in.ReferenceSelection.Files) > 8 {
			return in, inputError("最多追溯8个参考文件")
		}
		for _, f := range in.ReferenceSelection.Files {
			if len(f.Filename) > 1024 || len(f.ChapterIDs) > 2000 {
				return in, inputError("参考文件选择记录超出上限")
			}
			for _, id := range f.ChapterIDs {
				if len(id) > 96 {
					return in, inputError("参考章节编号无效")
				}
			}
		}
	}
	return in, nil
}
func (s *Service) StartLongImitation(ctx context.Context, owner string, input LongImitationInput) (out LongImitationResult, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	in, err := normalizeLongImitation(input)
	if err != nil {
		return out, err
	}
	digest := fingerprint(in)
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		receipt, e := tx.ReserveSubmission(ctx, owner, longImitationKind, in.IdempotencyKey, digest)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != digest {
			return conflict()
		}
		if len(receipt.Result) > 0 {
			return persistenceError(json.Unmarshal(receipt.Result, &out))
		}
		c := LongImitationCreation{Version: 1, Input: in, EffectiveReferenceText: in.ReferenceText, Statistics: ImitationStyleStatistics(in.ReferenceText, ""), StyleDimensions: ImitationStyleDimensions(), Styles: []importedStyleReference{}, Memos: []ProjectNote{}, AutoTitle: in.Title == "", CreatedAt: time.Now().UTC()}
		seen := map[string]bool{}
		styleCharacters := 0
		for _, id := range in.KnowledgeBookIDs {
			id = strings.TrimSpace(id)
			if id == "" || len(id) > 96 || seen[id] {
				return inputError("文风选择无效或重复")
			}
			seen[id] = true
			p, e := tx.GetStyleProfile(ctx, owner, id)
			if e != nil {
				return e
			}
			if strings.TrimSpace(p.Content) == "" {
				return inputError("文风内容为空")
			}
			styleCharacters += jsLength(p.Content)
			if styleCharacters > 200000 {
				return inputError("所选文风正文合计超过20万字元，请减少选择")
			}
			c.Styles = append(c.Styles, importedStyleReference{Profile: p, Digest: fingerprint(p)})
		}
		seen = map[string]bool{}
		for _, ref := range in.MemoReferences {
			key := memoryKey(ref.ProjectID, ref.NoteID)
			if ref.ProjectID == "" || ref.NoteID == "" || len(ref.ProjectID) > 96 || len(ref.NoteID) > 96 || seen[key] {
				return inputError("备忘录选择无效或重复")
			}
			seen[key] = true
			notes, e := tx.ListNotes(ctx, owner, ref.ProjectID)
			if e != nil {
				return e
			}
			found := false
			for _, n := range notes {
				if n.ID == ref.NoteID {
					found = true
					c.Memos = append(c.Memos, n)
					c.EffectiveReferenceText += "\n\n--- 备忘录: " + n.Title + " ---\n\n" + n.Content
					break
				}
			}
			if !found {
				return inputError("所选备忘录已不存在或不属于当前账号")
			}
		}
		if jsLength(c.EffectiveReferenceText) > LongImitationMaxUTF16 {
			return inputError("原文与备忘录合计超过200万个UTF16字元")
		}
		c.ReferenceChunks = imitationChunks(c.EffectiveReferenceText)
		c.ChunkingMethod = "50000_utf16_surrogate_safe"
		title := in.Title
		if c.AutoTitle {
			title = "待蓝图命名"
			c.ProvisionalTitle = title
		}
		project, e := tx.CreateLongImitationProject(ctx, owner, title, in.ModelID, in.TargetWordCount)
		if e != nil {
			return e
		}
		if e = putContent(ctx, tx, owner, project.ID, longImitationKind, "creation", "frozen", time.Time{}, c); e != nil {
			return e
		}
		session, e := tx.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: longImitationKind, Title: title})
		if e != nil {
			return e
		}
		child := NewService(tx, quietRuntime{s.runtime}, s.pipeline)
		child.promptCatalog = s.promptCatalog
		// Receipt keys support 128 characters; the existing run column supports
		// 96. Keep the external receipt untouched and bind a short internal key.
		internalKey := "long-imitation:" + fingerprint([]string{owner, longImitationKind, in.IdempotencyKey})
		run, e := child.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: novelprompt.LongImitationStartInstruction, ModelID: in.ModelID, ThinkingLevel: in.ThinkingLevel, IdempotencyKey: internalKey, ModeInputs: map[string]string{"referenceText": c.EffectiveReferenceText}})
		if e != nil {
			return e
		}
		out = LongImitationResult{project, session, run, c}
		raw, e := json.Marshal(out)
		if e != nil {
			return persistenceError(e)
		}
		return tx.CompleteSubmission(ctx, owner, longImitationKind, in.IdempotencyKey, submissionReceipt{Fingerprint: digest, ProjectID: project.ID, Result: raw})
	})
	if err == nil {
		s.Notify(ctx, owner, out.Run.ID)
	}
	return
}
func (s *Service) GetLongImitationRequest(ctx context.Context, owner, key string) (out LongImitationResult, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	if e := studioKey(key); e != nil {
		return out, e
	}
	r, ok, e := s.store.GetSubmissionReceipt(ctx, owner, longImitationKind, key)
	if e != nil {
		return out, e
	}
	if !ok || len(r.Result) == 0 {
		return out, &DomainError{Code: ErrProjectNotFound, Message: "未找到此仿写创建请求"}
	}
	err = json.Unmarshal(r.Result, &out)
	return out, persistenceError(err)
}
func loadLongImitation(ctx context.Context, st Store, owner, project string) (*LongImitationCreation, error) {
	var c LongImitationCreation
	ok, e := readContent(ctx, st, owner, project, longImitationKind, "creation", &c)
	if e != nil || !ok {
		return nil, e
	}
	return &c, nil
}
func (s *Service) GetLongImitation(ctx context.Context, owner, project string) (out LongImitationCreation, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	c, e := loadLongImitation(ctx, s.store, owner, project)
	if e != nil {
		return out, e
	}
	if c == nil {
		return out, &DomainError{Code: ErrProjectNotFound, Message: "此作品没有专用仿写配置"}
	}
	var analysis LongImitationAnalysis
	found, e := readContent(ctx, s.store, owner, project, longImitationKind, "analysis", &analysis)
	if e != nil {
		return out, e
	}
	if found {
		c.Analysis = &analysis
	}
	var requirements LongImitationRequirements
	found, e = readContent(ctx, s.store, owner, project, longImitationKind, "requirements", &requirements)
	if e != nil {
		return out, e
	}
	if found {
		c.Requirements = &requirements
	}
	return *c, nil
}

type ImitationMemoChoice struct {
	ProjectID string `json:"projectId"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content,omitempty"`
}

func (s *Service) LongImitationMemos(ctx context.Context, owner string) ([]ImitationMemoChoice, bool, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return nil, false, e
	}
	projects, e := s.store.ListProjects(ctx, owner)
	if e != nil {
		return nil, false, e
	}
	out := []ImitationMemoChoice{}
	for _, p := range projects {
		if projectReadOnly(p.Status) {
			continue
		}
		notes, e := s.store.ListNotes(ctx, owner, p.ID)
		if e != nil {
			return nil, false, e
		}
		for _, n := range notes {
			if len(out) >= 500 {
				return out, true, nil
			}
			out = append(out, ImitationMemoChoice{ProjectID: p.ID, ID: n.ID, Title: n.Title})
		}
	}
	return out, false, nil
}
