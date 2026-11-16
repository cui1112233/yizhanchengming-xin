package novelcreation

import (
	"context"
	"strings"
	"time"
)

func projectReadOnly(status string) bool { return status == "archived" || status == "purging" }
func archivedError() error {
	return inputError("作品已归档或正在清理，不能修改；清理中的作品不可恢复")
}
func (s *Service) ConfigureRetention(archive, restore func(context.Context, string, string) error) {
	s.archiveProject = archive
	s.restoreProject = restore
}
func (s *Service) ArchiveProject(ctx context.Context, owner, project string) error {
	if s.archiveProject != nil {
		return s.archiveProject(ctx, owner, project)
	}
	if st, ok := s.store.(*MemoryStore); ok {
		return st.archiveMemory(ctx, owner, project, false)
	}
	return inputError("作品归档服务未配置")
}
func (s *Service) RestoreProject(ctx context.Context, owner, project string) error {
	if s.restoreProject != nil {
		return s.restoreProject(ctx, owner, project)
	}
	if st, ok := s.store.(*MemoryStore); ok {
		return st.archiveMemory(ctx, owner, project, true)
	}
	return inputError("作品恢复服务未配置")
}
func (s *MemoryStore) archiveMemory(_ context.Context, owner, project string, restore bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, project)
	p, ok := s.projects[key]
	if !ok || p.DeletedAt != nil {
		return inputError("作品不存在")
	}
	if p.Status == "purging" {
		return inputError("作品正在清理，不能恢复或修改")
	}
	if restore {
		if p.Status != "archived" {
			return inputError("作品未归档")
		}
		p.Status = "draft"
		p.UpdatedAt = time.Now().UTC()
		s.projects[key] = p
		return nil
	}
	for k, r := range s.runs {
		if r.ProjectID == project && strings.HasPrefix(k, memoryKey(owner)+"\x00") && !isTerminalRunStatus(r.Status) && r.Status != RunPaused && r.Status != RunAwaitingConfirmation {
			return inputError("请先停止正在运行或等待派发的任务")
		}
	}
	for _, r := range s.contentRecords {
		if r.Owner == owner && r.ProjectID == project && ((r.Kind == "document" && (r.Status == "queued" || r.Status == "parsing")) || ((r.Kind == "document" || r.Kind == "export") && r.Status == "uploading" && r.LeaseUntil.After(time.Now().UTC()))) {
			return inputError("请等待资料上传解析结束")
		}
	}
	p.Status = "archived"
	s.projects[key] = p
	for k, v := range s.schedules {
		if v.ProjectID == project && v.OwnerUsername == owner {
			v.Enabled = false
			v.AutomationState.Status = "paused"
			v.StoppedReason = "作品已归档，恢复后请手动继续"
			s.schedules[k] = v
		}
	}
	for k, v := range s.batches {
		if v.ProjectID == project && v.OwnerUsername == owner {
			v.Status = BatchPaused
			v.AutomationState.Status = "paused"
			s.batches[k] = v
		}
	}
	return nil
}
func (s *Service) PurgeDocumentObjects(ctx context.Context, owner, project string) error {
	for _, kind := range []string{"document", "export"} {
		rows, e := s.store.ListContentRecords(ctx, owner, project, kind)
		if e != nil {
			return e
		}
		for _, r := range rows {
			var o DocumentObject
			if kind == "document" {
				var v documentStored
				if e = decodeRecord(r.Data, &v); e != nil {
					return e
				}
				o = DocumentObject(v.Object)
			} else {
				var v exportRecord
				if e = decodeRecord(r.Data, &v); e != nil {
					return e
				}
				o = DocumentObject(v.Object)
			}
			if o.Key != "" {
				if s.documents == nil {
					return inputError("未配置TOS，无法清理原件；归档记录已保留")
				}
				if e = s.documents.Delete(ctx, o); e != nil {
					return e
				}
			}
		}
	}
	return nil
}

// Existing project mutators participate in the same project lock as archive,
// claims and final commits. Reads remain available in the archive.
func mutateValue[T any](ctx context.Context, s *Service, owner, project string, fn func(*Service) (T, error)) (out T, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error { child := *s; child.store = tx; var e error; out, e = fn(&child); return e })
	return
}
func (s *Service) CreateSession(ctx context.Context, o string, in CreateSessionInput) (Session, error) {
	return mutateValue(ctx, s, o, in.ProjectID, func(c *Service) (Session, error) { return c.createSession(ctx, o, in) })
}
func (s *Service) UpdateProjectSettings(ctx context.Context, o, p string, in ProjectSettings) (ProjectSettings, error) {
	return mutateValue(ctx, s, o, p, func(c *Service) (ProjectSettings, error) { return c.updateProjectSettings(ctx, o, p, in) })
}
func (s *Service) CreateMaterial(ctx context.Context, o, p string, in CreateMaterialInput) (Material, error) {
	return mutateValue(ctx, s, o, p, func(c *Service) (Material, error) { return c.createMaterial(ctx, o, p, in) })
}
func (s *Service) DeleteMaterial(ctx context.Context, o, p, id string) error {
	_, e := mutateValue(ctx, s, o, p, func(c *Service) (bool, error) {
		if e := c.deleteMaterial(ctx, o, p, id); e != nil {
			return false, e
		}
		return true, c.store.TouchProject(ctx, o, p)
	})
	return e
}
func (s *Service) CreateNote(ctx context.Context, o, p string, in ProjectNoteInput) (ProjectNote, error) {
	return mutateValue(ctx, s, o, p, func(c *Service) (ProjectNote, error) { return c.createNote(ctx, o, p, in) })
}
func (s *Service) UpdateNote(ctx context.Context, o, p, id string, in ProjectNoteInput) (ProjectNote, error) {
	return mutateValue(ctx, s, o, p, func(c *Service) (ProjectNote, error) { return c.updateNote(ctx, o, p, id, in) })
}
func (s *Service) DeleteNote(ctx context.Context, o, p, id string) error {
	_, e := mutateValue(ctx, s, o, p, func(c *Service) (bool, error) {
		if e := c.deleteNote(ctx, o, p, id); e != nil {
			return false, e
		}
		return true, c.store.TouchProject(ctx, o, p)
	})
	return e
}
func (s *Service) CreateSchedule(ctx context.Context, o, p string, in ScheduleInput) (Schedule, error) {
	return mutateValue(ctx, s, o, p, func(c *Service) (Schedule, error) { return c.createSchedule(ctx, o, p, in) })
}
func (s *Service) DeleteSchedule(ctx context.Context, o, p, id string) error {
	_, e := mutateValue(ctx, s, o, p, func(c *Service) (bool, error) { return true, c.deleteSchedule(ctx, o, p, id) })
	return e
}
