package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
)

type PreviousChapter struct {
	ChapterID string `json:"chapter_id"`
	Version   int    `json:"version"`
	Number    int    `json:"number"`
	Ending    string `json:"ending"`
}
type chapterContinuity struct {
	Truth    json.RawMessage
	Memory   json.RawMessage
	Chapters []PreviousChapter
}

func continuityError(cause error) error {
	return &DomainError{Code: ErrPersistenceFailed, Message: "前文资料读取失败，已停止本次创作", Cause: cause}
}
func validContinuityJSON(raw json.RawMessage) bool {
	var v map[string]any
	return json.Unmarshal(raw, &v) == nil && len(v) > 0
}
func sameJSON(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}

// Bind a memory to the exact saved chapter version and its generating run.
func bindChapterMemory(memory Memory, version int, runtime json.RawMessage) (Memory, bool, error) {
	var saved struct {
		Truth  json.RawMessage `json:"truthSettlement"`
		Memory json.RawMessage `json:"memoryIndex"`
	}
	if json.Unmarshal(runtime, &saved) != nil || !validContinuityJSON(saved.Truth) || !validContinuityJSON(saved.Memory) || !sameJSON(memory.Snapshot, saved.Memory) {
		return Memory{}, false, continuityError(nil)
	}
	memory.ChapterVersion = version
	memory.PreviousTruth = cloneRawJSON(saved.Truth)
	memory.Snapshot = cloneRawJSON(memory.Snapshot)
	return memory, true, nil
}
func (s *MemoryStore) GetChapterMemory(_ context.Context, owner, projectID, chapterID string) (Memory, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || p.DeletedAt != nil {
		return Memory{}, false, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	b, ok := s.chapterBundles[memoryKey(owner, projectID, chapterID)]
	if !ok {
		return Memory{}, false, nil
	}
	if b.ChapterVersion.Version != b.Chapter.Version || b.ChapterVersion.Content != b.Chapter.Content || b.ChapterVersion.ChapterID != chapterID {
		return Memory{}, false, continuityError(nil)
	}
	r, ok := s.runs[memoryKey(owner, b.ChapterVersion.RunID)]
	if !ok || r.ProjectID != projectID {
		return Memory{}, false, continuityError(nil)
	}
	var newest Memory
	for _, m := range b.Memories {
		if (m.ProjectID == "" || m.ProjectID == projectID) && (m.ChapterID == "" || m.ChapterID == chapterID) && (newest.ID == "" || m.UpdatedAt.After(newest.UpdatedAt)) {
			newest = m
		}
	}
	if newest.ID == "" {
		return Memory{}, false, nil
	}
	newest.ProjectID = projectID
	newest.ChapterID = chapterID
	return bindChapterMemory(newest, b.ChapterVersion.Version, r.RuntimeSnapshot)
}
func (s *MySQLStore) GetChapterMemory(ctx context.Context, owner, projectID, chapterID string) (Memory, bool, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Memory{}, false, err
	}
	if _, err = s.GetProject(ctx, owner, projectID); err != nil {
		return Memory{}, false, err
	}
	var m Memory
	var runtime []byte
	var raw []byte
	var version int
	err = s.db.QueryRowContext(ctx, `SELECT m.memory_id,m.project_id,m.chapter_id,m.snapshot_json,m.created_at,m.updated_at,v.version_no,r.runtime_snapshot_json
 FROM novel_memories m
 JOIN novel_chapters c ON c.owner_username=m.owner_username AND c.project_id=m.project_id AND c.chapter_id=m.chapter_id
 JOIN novel_chapter_versions v ON v.owner_username=c.owner_username AND v.project_id=c.project_id AND v.chapter_id=c.chapter_id AND v.version_no=c.version_no AND BINARY v.content=BINARY c.content
 JOIN novel_runs r ON r.owner_username=v.owner_username AND r.project_id=v.project_id AND r.run_id=v.run_id
 JOIN novel_projects p ON p.owner_username=c.owner_username AND p.project_id=c.project_id AND p.deleted_at IS NULL
 WHERE m.owner_username=? AND m.project_id=? AND m.chapter_id=?
 ORDER BY m.updated_at DESC,m.memory_id DESC,v.created_at DESC LIMIT 1`, owner, strings.TrimSpace(projectID), strings.TrimSpace(chapterID)).Scan(&m.ID, &m.ProjectID, &m.ChapterID, &raw, &m.CreatedAt, &m.UpdatedAt, &version, &runtime)
	if errors.Is(err, sql.ErrNoRows) {
		return Memory{}, false, nil
	}
	if err != nil {
		return Memory{}, false, persistenceError(err)
	}
	m.Snapshot = raw
	return bindChapterMemory(m, version, runtime)
}
func (s *Service) loadContinuity(ctx context.Context, owner, projectID string, chapters []Chapter) (chapterContinuity, error) {
	out := chapterContinuity{Truth: json.RawMessage(`{}`), Memory: json.RawMessage(`{}`), Chapters: []PreviousChapter{}}
	if len(chapters) == 0 {
		return out, nil
	}
	ordered := append([]Chapter(nil), chapters...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })
	latest := ordered[len(ordered)-1]
	memory, found, err := s.store.GetChapterMemory(ctx, owner, projectID, latest.ID)
	if err != nil {
		return out, continuityError(err)
	}
	if !found || memory.ChapterVersion != latest.Version || !validContinuityJSON(memory.PreviousTruth) || !validContinuityJSON(memory.Snapshot) {
		return out, continuityError(nil)
	}
	out.Truth = memory.PreviousTruth
	out.Memory = memory.Snapshot
	// The persisted memory carries older summaries. Keep the immediate ending intact.
	ending := []rune(strings.TrimSpace(latest.Content))
	if len(ending) > 2000 {
		ending = ending[len(ending)-2000:]
	}
	out.Chapters = append(out.Chapters, PreviousChapter{ChapterID: latest.ID, Version: latest.Version, Number: latest.Number, Ending: string(ending)})
	return out, nil
}
func (s *Service) validateContinuity(ctx context.Context, owner, projectID string, previous []PreviousChapter, number int, retry bool) error {
	chapters, err := s.store.ListChapters(ctx, owner, projectID)
	if err != nil {
		return continuityError(err)
	}
	if retry {
		before := make([]Chapter, 0, len(chapters))
		for _, chapter := range chapters {
			if chapter.Number < number {
				before = append(before, chapter)
			}
		}
		chapters = before
	}
	if number <= 1 && len(chapters) == 0 {
		return nil
	}
	if len(previous) == 0 {
		return continuityError(nil)
	}
	current, err := s.loadContinuity(ctx, owner, projectID, chapters)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(previous, current.Chapters) {
		return continuityError(nil)
	}
	return nil
}
func (s *Service) GetSessionSnapshot(ctx context.Context, owner, sessionID string) (Snapshot, error) {
	if s == nil || s.store == nil {
		return Snapshot{}, continuityError(nil)
	}
	owner = strings.TrimSpace(owner)
	session, err := s.store.GetSession(ctx, owner, strings.TrimSpace(sessionID))
	if err != nil {
		return Snapshot{}, err
	}
	project, err := s.store.GetProject(ctx, owner, session.ProjectID)
	if err != nil {
		return Snapshot{}, err
	}
	messages, err := s.store.ListMessages(ctx, owner, session.ID, 0)
	if err != nil {
		return Snapshot{}, err
	}
	chapters, err := s.store.ListChapters(ctx, owner, project.ID)
	if err != nil {
		return Snapshot{}, err
	}
	var latest Run
	if session.CurrentRunID != "" {
		latest, err = s.store.GetRun(ctx, owner, session.CurrentRunID)
		if err != nil {
			return Snapshot{}, err
		}
		if latest.SessionID != session.ID || latest.ProjectID != project.ID {
			return Snapshot{}, continuityError(nil)
		}
	} else {
		runs, listErr := s.store.ListRuns(ctx, owner, project.ID)
		if listErr != nil {
			return Snapshot{}, listErr
		}
		for _, r := range runs {
			if r.SessionID == session.ID && (latest.ID == "" || r.CreatedAt.After(latest.CreatedAt)) {
				latest = r
			}
		}
	}
	events := []RunEvent{}
	if latest.ID != "" {
		events, err = s.store.ListRunEvents(ctx, owner, latest.ID, 0)
		if err != nil {
			return Snapshot{}, err
		}
	}
	if messages == nil {
		messages = []Message{}
	}
	if chapters == nil {
		chapters = []Chapter{}
	}
	executionState := ""
	if latest.ID != "" {
		execution, _, getErr := s.store.GetExecution(ctx, owner, latest.ID)
		if getErr != nil {
			return Snapshot{}, getErr
		}
		executionState = execution.State
	}
	return Snapshot{ExecutionState: executionState, Project: project, Session: session, Run: latest, Messages: messages, Chapters: chapters, Events: events}, nil
}

func (s *Service) validateTargetVersion(ctx context.Context, owner, projectID string, target *PreviousChapter) error {
	if target == nil {
		return nil
	}
	chapters, err := s.store.ListChapters(ctx, owner, projectID)
	if err != nil {
		return continuityError(err)
	}
	if target.Version == 0 {
		for _, chapter := range chapters {
			if chapter.Number == target.Number {
				return continuityError(nil)
			}
		}
		return nil
	}
	for _, chapter := range chapters {
		if chapter.ID == target.ChapterID && chapter.Version == target.Version && chapter.Content == target.Ending {
			return nil
		}
	}
	return continuityError(nil)
}
