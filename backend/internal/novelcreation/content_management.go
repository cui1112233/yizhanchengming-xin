package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

type ContentManagementStore interface {
	TouchProject(context.Context, string, string) error
	RenameProject(context.Context, string, string, string) (Project, error)
	UpdateSession(context.Context, string, Session) (Session, error)
	WriteManualChapter(context.Context, string, Chapter, ChapterVersion) error
	ListChapterVersions(context.Context, string, string, string) ([]ChapterVersion, error)
}
type ChapterEdit struct {
	ExpectedVersion int    `json:"expectedVersion"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	RestoreVersion  int    `json:"-"`
}
type ChapterSource struct {
	ID      string `json:"id"`
	Number  int    `json:"number"`
	Version int    `json:"version"`
	Digest  string `json:"digest"`
}
type ContinuityEntry struct {
	Source ChapterSource   `json:"source"`
	Truth  json.RawMessage `json:"truth"`
	Memory json.RawMessage `json:"memory"`
}
type ContinuityState struct {
	DirtyFrom       int               `json:"dirtyFrom"`
	Revision        int               `json:"revision"`
	AnalysisVersion int               `json:"analysisVersion"`
	RunID           string            `json:"runId,omitempty"`
	Sources         []ChapterSource   `json:"sources"`
	Entries         []ContinuityEntry `json:"entries"`
}

func chapterSources(ch []Chapter) []ChapterSource {
	out := make([]ChapterSource, 0, len(ch))
	for _, c := range ch {
		out = append(out, ChapterSource{c.ID, c.Number, c.Version, fingerprint(c.Content)})
	}
	return out
}
func (s *Service) GetContinuityState(ctx context.Context, owner, project string) (v ContinuityState, err error) {
	_, err = readContent(ctx, s.store, owner, project, "continuity", "current", &v)
	return
}
func (s *Service) ListChapterVersions(ctx context.Context, owner, project, id string) ([]ChapterVersion, error) {
	return s.store.ListChapterVersions(ctx, owner, project, id)
}
func (s *Service) EditChapter(ctx context.Context, owner, project, id string, in ChapterEdit) (out Chapter, err error) {
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" || utf8.RuneCountInString(in.Content) > 200000 {
		return out, inputError("正文不能为空且不能超过20万字元")
	}
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		chapters, e := tx.ListChapters(ctx, owner, project)
		if e != nil {
			return e
		}
		found := false
		for _, c := range chapters {
			if c.ID == id {
				out = c
				found = true
				break
			}
		}
		if !found {
			return inputError("章节不存在")
		}
		if out.Version != in.ExpectedVersion {
			return inputError("章节已改版，请刷新后再编辑")
		}
		out.Version++
		out.Content = in.Content
		if strings.TrimSpace(in.Title) != "" {
			out.Title = strings.TrimSpace(in.Title)
		}
		out.WordCount = utf8.RuneCountInString(out.Content)
		out.AuditStatus = "manual"
		out.UpdatedAt = time.Now().UTC()
		source := "manual_edit"
		if in.RestoreVersion > 0 {
			source = "restore"
		}
		meta, _ := json.Marshal(map[string]any{"source": source, "restoreVersion": in.RestoreVersion, "title": out.Title})
		version := ChapterVersion{ID: novelID("version"), ChapterID: id, Version: out.Version, Content: out.Content, Snapshot: meta, CreatedAt: out.UpdatedAt}
		if e = tx.WriteManualChapter(ctx, owner, out, version); e != nil {
			return e
		}
		var state ContinuityState
		if _, e = readContent(ctx, tx, owner, project, "continuity", "current", &state); e != nil {
			return e
		}
		if state.DirtyFrom == 0 || out.Number < state.DirtyFrom {
			state.DirtyFrom = out.Number
		}
		state.Revision++
		return putContent(ctx, tx, owner, project, "continuity", "current", "dirty", time.Time{}, state)
	})
	return
}
func (s *Service) RestoreChapter(ctx context.Context, owner, project, id string, expected, version int) (Chapter, error) {
	versions, e := s.store.ListChapterVersions(ctx, owner, project, id)
	if e != nil {
		return Chapter{}, e
	}
	for _, v := range versions {
		if v.Version == version {
			var meta struct {
				Title string `json:"title"`
			}
			_ = json.Unmarshal(v.Snapshot, &meta)
			return s.EditChapter(ctx, owner, project, id, ChapterEdit{ExpectedVersion: expected, Title: meta.Title, Content: v.Content, RestoreVersion: version})
		}
	}
	return Chapter{}, inputError("历史版本不存在")
}
func (s *Service) RenameProject(ctx context.Context, owner, project, title string) (out Project, err error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 255 {
		return out, inputError("请输入1到255字元的作品名称")
	}
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error { var e error; out, e = tx.RenameProject(ctx, owner, project, title); return e })
	return
}
func (s *Service) ManageSession(ctx context.Context, owner, project, id, action, title string) (out Session, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		var e error
		out, e = tx.GetSession(ctx, owner, id)
		if e != nil {
			return e
		}
		if out.ProjectID != project {
			return inputError("会话不属于当前作品")
		}
		switch action {
		case "rename":
			title = strings.TrimSpace(title)
			if title == "" || utf8.RuneCountInString(title) > 255 {
				return inputError("请输入有效会话名称")
			}
			out.Title = title
		case "archive":
			runs, e := tx.ListRuns(ctx, owner, project)
			if e != nil {
				return e
			}
			for _, r := range runs {
				if r.SessionID == id && !isTerminalRunStatus(r.Status) && r.Status != RunPaused && r.Status != RunAwaitingConfirmation {
					return inputError("请先停止会话中运行的任务")
				}
			}
			now := time.Now().UTC()
			out.ArchivedAt = &now
			schedules, e := tx.ListSchedules(ctx, owner, project)
			if e != nil {
				return e
			}
			for _, plan := range schedules {
				if plan.SessionID == id {
					plan.Enabled = false
					plan.AutomationState.Status = "paused"
					plan.StoppedReason = "会话已归档，恢复后请手动继续"
					if _, e = tx.UpdateSchedule(ctx, owner, plan); e != nil {
						return e
					}
				}
			}
			batches, e := tx.ListBatches(ctx, owner, project)
			if e != nil {
				return e
			}
			for _, plan := range batches {
				if plan.SessionID == id && (plan.Status == BatchRunning || plan.Status == BatchPaused) {
					plan.Status = BatchPaused
					plan.AutomationState.Status = "paused"
					if _, e = tx.UpdateBatch(ctx, owner, plan); e != nil {
						return e
					}
				}
			}

		case "restore":
			out.ArchivedAt = nil
		default:
			return inputError("不支持的会话操作")
		}
		out, e = tx.UpdateSession(ctx, owner, out)
		return e
	})
	return
}
