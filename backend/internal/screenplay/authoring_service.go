package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrAccessDenied = errors.New("没有此项目的操作权限")

type Chapter struct {
	PlanningPromptID      string                  `json:"planning_prompt_id,omitempty"`
	PlanningPromptDigest  string                  `json:"planning_prompt_digest,omitempty"`
	PlanningConfigDigest  string                  `json:"planning_config_digest,omitempty"`
	PlanningRequestKey    string                  `json:"planning_request_key,omitempty"`
	PlanningRequestDigest string                  `json:"planning_request_digest,omitempty"`
	PlanningStatus        string                  `json:"planning_status,omitempty"`
	PlanningModel         string                  `json:"planning_model,omitempty"`
	PlanningText          string                  `json:"planning_text,omitempty"`
	PlanningCoverage      *StoryboardPlanCoverage `json:"planning_coverage,omitempty"`
	PlanDensity           string                  `json:"plan_density,omitempty"`
	PlanningSourceSHA256  string                  `json:"planning_source_sha256,omitempty"`
	PlanningUseOriginal   bool                    `json:"planning_use_original,omitempty"`
	RawResult             string                  `json:"raw_result,omitempty"`
	ID                    string                  `json:"id"`
	Title                 string                  `json:"title"`
	SourceText            string                  `json:"source_text"`
	ScriptText            string                  `json:"script_text"`
}
type MemberGrant struct {
	Username string `json:"username"`
	Read     bool   `json:"read"`
	Edit     bool   `json:"edit"`
	Generate bool   `json:"generate"`
}
type AssistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type AuthoringAction struct {
	Kind             string   `json:"kind"`
	TargetID         string   `json:"target_id,omitempty"`
	Content          string   `json:"content,omitempty"`
	Chapter          *Chapter `json:"chapter,omitempty"`
	Asset            *Asset   `json:"asset,omitempty"`
	Shot             *Shot    `json:"shot,omitempty"`
	ExpectedRevision int64    `json:"expected_revision"`
}
type AssistantProposal struct {
	ID       string            `json:"id"`
	Summary  string            `json:"summary"`
	Actions  []AuthoringAction `json:"actions"`
	Status   string            `json:"status"`
	Revision int64             `json:"revision"`
}
type AuthoringState struct {
	ProjectID string              `json:"project_id"`
	Owner     string              `json:"owner"`
	Revision  int64               `json:"revision"`
	Chapters  []Chapter           `json:"chapters"`
	Display   map[string]any      `json:"display"`
	Messages  []AssistantMessage  `json:"messages"`
	Proposals []AssistantProposal `json:"proposals"`
	Members   []MemberGrant       `json:"members"`
	Archived  bool                `json:"archived"`
}
type AuthoringRepository interface {
	Load(context.Context, string, string) (AuthoringState, error)
	Save(context.Context, AuthoringState, int64) (AuthoringState, error)
	List(context.Context) ([]AuthoringState, error)
}
type MemoryAuthoringRepository struct {
	mu   sync.Mutex
	rows map[string]AuthoringState
}

func NewMemoryAuthoringRepository() *MemoryAuthoringRepository {
	return &MemoryAuthoringRepository{rows: map[string]AuthoringState{}}
}
func cloneAuthoring(s AuthoringState) AuthoringState {
	b, _ := json.Marshal(s)
	var out AuthoringState
	json.Unmarshal(b, &out)
	return out
}
func (r *MemoryAuthoringRepository) Load(_ context.Context, o, p string) (AuthoringState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.rows[projectKey(o, p)]
	if !ok {
		return AuthoringState{}, ErrNotFound
	}
	return cloneAuthoring(s), nil
}
func (r *MemoryAuthoringRepository) Save(_ context.Context, s AuthoringState, v int64) (AuthoringState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old := r.rows[projectKey(s.Owner, s.ProjectID)]
	if old.Revision != v {
		return AuthoringState{}, ErrRevisionConflict
	}
	s.Revision = v + 1
	r.rows[projectKey(s.Owner, s.ProjectID)] = cloneAuthoring(s)
	return s, nil
}
func (r *MemoryAuthoringRepository) List(_ context.Context) ([]AuthoringState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []AuthoringState{}
	for _, s := range r.rows {
		out = append(out, cloneAuthoring(s))
	}
	return out, nil
}

type AuthoringService struct {
	Store      Store
	Repository AuthoringRepository
	Generator  TextGenerator
}

func (s AuthoringService) Load(ctx context.Context, o, p string) (AuthoringState, error) {
	if _, e := s.Store.GetProject(ctx, o, p); e != nil {
		return AuthoringState{}, e
	}
	st, e := s.Repository.Load(ctx, o, p)
	if errors.Is(e, ErrNotFound) {
		return AuthoringState{ProjectID: p, Owner: o, Chapters: []Chapter{}, Messages: []AssistantMessage{}, Proposals: []AssistantProposal{}, Members: []MemberGrant{}, Display: map[string]any{"editor_version": "studio", "active_tab": "source", "font_size": 14}}, nil
	}
	return st, e
}
func (s AuthoringService) Save(ctx context.Context, o, p string, st AuthoringState, v int64) (AuthoringState, error) {
	if _, e := s.Store.GetProject(ctx, o, p); e != nil {
		return AuthoringState{}, e
	}
	seen := map[string]bool{}
	for _, c := range st.Chapters {
		if c.ID == "" || seen[c.ID] || strings.TrimSpace(c.Title) == "" {
			return AuthoringState{}, fmt.Errorf("%w: 章节编号必须唯一且标题不能为空", ErrInvalidArgument)
		}
		seen[c.ID] = true
	}
	for i := range st.Proposals {
		if st.Proposals[i].Status == "pending" && st.Proposals[i].Revision != v+1 {
			st.Proposals[i].Status = "stale"
		}
	}
	st.ProjectID = p
	st.Owner = o
	return s.Repository.Save(ctx, st, v)
}
func (s AuthoringService) Resolve(ctx context.Context, actor, p, permission string) (string, error) {
	if _, e := s.Store.GetProject(ctx, actor, p); e == nil {
		if permission == "generate" {
			st, err := s.Repository.Load(ctx, actor, p)
			if err != nil && !errors.Is(err, ErrNotFound) {
				return "", err
			}
			if err == nil && st.Archived {
				return "", fmt.Errorf("%w: 项目已归档，请先恢复作品", ErrInvalidState)
			}
		}
		return actor, nil
	}
	rows, e := s.Repository.List(ctx)
	if e != nil {
		return "", e
	}
	for _, st := range rows {
		if st.ProjectID != p {
			continue
		}
		for _, g := range st.Members {
			if g.Username == actor && g.Read && (permission == "read" || permission == "edit" && g.Edit || permission == "generate" && g.Generate) {
				if _, e = s.Store.GetProject(ctx, st.Owner, p); e == nil {
					if permission == "generate" && st.Archived {
						return "", fmt.Errorf("%w: 项目已归档，请先恢复作品", ErrInvalidState)
					}
					return st.Owner, nil
				}
			}
		}
	}
	return "", ErrAccessDenied
}
func ValidateAuthoringActions(actions []AuthoringAction) error {
	if len(actions) == 0 || len(actions) > 20 {
		return fmt.Errorf("%w: 编辑动作数量不合适", ErrInvalidArgument)
	}
	for _, a := range actions {
		switch a.Kind {
		case "document":
			if strings.TrimSpace(a.Content) == "" {
				return ErrEmptyContent
			}
		case "chapter":
			if a.Chapter == nil || a.Chapter.ID == "" || a.Chapter.Title == "" {
				return ErrInvalidArgument
			}
		case "asset":
			if a.Asset == nil || a.TargetID == "" {
				return ErrInvalidArgument
			}
		case "shot":
			if a.Shot == nil || a.TargetID == "" {
				return ErrInvalidArgument
			}
		default:
			return fmt.Errorf("%w: 不支持的助手动作 %s", ErrInvalidArgument, a.Kind)
		}
	}
	return nil
}
func (s AuthoringService) WithLock(ctx context.Context, o, p string, fn func() error) error {
	token := newScreenplayID("authoring")
	name := "authoring:" + o + ":" + p
	if e := s.Store.AcquireLock(ctx, name, token, 10*time.Minute); e != nil {
		return e
	}
	defer s.Store.ReleaseLock(context.Background(), name, token)
	return fn()
}
func (s AuthoringService) Apply(ctx context.Context, o, p, id string, v int64, confirmed bool) (AuthoringState, error) {
	var out AuthoringState
	if !confirmed {
		return out, fmt.Errorf("%w: 请确认编辑内容", ErrInvalidArgument)
	}
	e := s.WithLock(ctx, o, p, func() error {
		st, e := s.Load(ctx, o, p)
		if e != nil {
			return e
		}
		if st.Revision != v {
			return ErrRevisionConflict
		}
		index := -1
		for i, x := range st.Proposals {
			if x.ID == id {
				index = i
			}
		}
		if index < 0 {
			return ErrNotFound
		}
		proposal := st.Proposals[index]
		if proposal.Status != "pending" || proposal.Revision != v {
			return ErrRevisionConflict
		}
		if e = ValidateAuthoringActions(proposal.Actions); e != nil {
			return e
		}
		if len(proposal.Actions) != 1 {
			return fmt.Errorf("%w: 每次只能确认一个编辑动作", ErrInvalidArgument)
		}
		a := proposal.Actions[0]
		switch a.Kind {
		case "chapter":
			found := false
			for i, c := range st.Chapters {
				if c.ID == a.Chapter.ID {
					st.Chapters[i] = *a.Chapter
					found = true
				}
			}
			if !found {
				return ErrNotFound
			}
		case "document":
			doc, e := s.Store.ActiveDocument(ctx, o, p)
			if e != nil && !errors.Is(e, ErrNotFound) {
				return e
			}
			if doc.Revision != a.ExpectedRevision {
				return ErrRevisionConflict
			}
			if _, e = PublishChapterScripts(ctx, s.Store, o, p, a.Content); e != nil {
				return e
			}
		case "asset":
			if _, e = s.Store.UpdateAsset(ctx, o, p, a.TargetID, *a.Asset, a.ExpectedRevision); e != nil {
				return e
			}
		case "shot":
			doc, e := s.Store.ActiveDocument(ctx, o, p)
			if e != nil {
				return e
			}
			if doc.Revision != a.ExpectedRevision {
				return ErrRevisionConflict
			}
			shots, e := s.Store.ListShots(ctx, o, p, doc.ID)
			if e != nil {
				return e
			}
			found := false
			for _, sh := range shots {
				if sh.ID == a.TargetID {
					found = true
				}
			}
			if !found {
				return ErrNotFound
			}
			if _, e = s.Store.UpdateShot(ctx, o, p, a.TargetID, *a.Shot); e != nil {
				return e
			}
		}
		if a.Kind == "chapter" {
			for _, c := range st.Chapters {
				if strings.TrimSpace(c.ScriptText) != "" {
					if _, e = PublishChapterCollection(ctx, s.Store, o, p, st.Chapters); e != nil {
						return e
					}
					break
				}
			}
		}
		st.Proposals[index].Status = "applied"
		out, e = s.Save(ctx, o, p, st, v)
		if e == nil {
			e = InvalidateAuthoringContent(ctx, s.Store, o, p, a.Kind)
		}
		return e
	})
	return out, e
}
