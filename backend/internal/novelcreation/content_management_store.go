package novelcreation

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *MemoryStore) RenameProject(ctx context.Context, owner, id, title string) (Project, error) {
	p, e := s.GetProject(ctx, owner, id)
	if e != nil {
		return p, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.Title = title
	p.UpdatedAt = time.Now().UTC()
	s.projects[memoryKey(owner, id)] = p
	return p, nil
}
func (s *MemoryStore) UpdateSession(ctx context.Context, owner string, v Session) (Session, error) {
	old, e := s.GetSession(ctx, owner, v.ID)
	if e != nil {
		return v, e
	}
	if old.ProjectID != v.ProjectID {
		return v, inputError("会话作品不匹配")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v.UpdatedAt = time.Now().UTC()
	s.sessions[memoryKey(owner, v.ID)] = v
	return v, nil
}
func (s *MemoryStore) ListChapterVersions(ctx context.Context, owner, project, id string) ([]ChapterVersion, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return nil, e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.chapterBundles[memoryKey(owner, project, id)]; !ok {
		return nil, inputError("章节不存在")
	}
	out := []ChapterVersion{}
	for k, v := range s.chapterVersions {
		if strings.HasPrefix(k, memoryKey(owner, project, id)+"\x00") {
			v.Snapshot = cloneRawJSON(v.Snapshot)
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}
func (s *MemoryStore) WriteManualChapter(_ context.Context, owner string, c Chapter, v ChapterVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, c.ProjectID, c.ID)
	b, ok := s.chapterBundles[key]
	if !ok || b.Chapter.Version+1 != c.Version {
		return conflict()
	}
	b = cloneChapterBundle(b)
	b.Chapter = c
	b.ChapterVersion = v
	s.chapterBundles[key] = b
	s.chapterVersions[memoryKey(owner, c.ProjectID, c.ID, fmtVersion(v.Version))] = v
	return nil
}
func fmtVersion(n int) string { return fmt.Sprintf("%d", n) }
func (s *MySQLStore) RenameProject(ctx context.Context, owner, id, title string) (Project, error) {
	_, e := s.db.ExecContext(ctx, `UPDATE novel_projects SET title=?,updated_at=UTC_TIMESTAMP(6) WHERE owner_username=? AND project_id=? AND deleted_at IS NULL`, title, owner, id)
	if e != nil {
		return Project{}, persistenceError(e)
	}
	return s.GetProject(ctx, owner, id)
}
func (s *MySQLStore) UpdateSession(ctx context.Context, owner string, v Session) (Session, error) {
	r, e := s.db.ExecContext(ctx, `UPDATE novel_sessions SET title=?,archived_at=?,updated_at=UTC_TIMESTAMP(6) WHERE owner_username=? AND project_id=? AND session_id=?`, v.Title, v.ArchivedAt, owner, v.ProjectID, v.ID)
	if e != nil {
		return v, persistenceError(e)
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return v, inputError("会话不存在")
	}
	return s.GetSession(ctx, owner, v.ID)
}
func (s *MySQLStore) ListChapterVersions(ctx context.Context, owner, project, id string) ([]ChapterVersion, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT version_id,chapter_id,COALESCE(run_id,''),version_no,content,snapshot_json,created_at FROM novel_chapter_versions WHERE owner_username=? AND project_id=? AND chapter_id=? ORDER BY version_no`, owner, project, id)
	if e != nil {
		return nil, persistenceError(e)
	}
	defer rows.Close()
	out := []ChapterVersion{}
	for rows.Next() {
		var v ChapterVersion
		if e = rows.Scan(&v.ID, &v.ChapterID, &v.RunID, &v.Version, &v.Content, &v.Snapshot, &v.CreatedAt); e != nil {
			return nil, persistenceError(e)
		}
		out = append(out, v)
	}
	return out, persistenceError(rows.Err())
}
func (s *MySQLStore) WriteManualChapter(ctx context.Context, owner string, c Chapter, v ChapterVersion) error {
	r, e := s.db.ExecContext(ctx, `UPDATE novel_chapters SET title=?,content=?,word_count=?,audit_status='manual',version_no=?,updated_at=? WHERE owner_username=? AND project_id=? AND chapter_id=? AND version_no=?`, c.Title, c.Content, c.WordCount, c.Version, c.UpdatedAt, owner, c.ProjectID, c.ID, c.Version-1)
	if e != nil {
		return persistenceError(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return conflict()
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO novel_chapter_versions (version_id,owner_username,project_id,chapter_id,run_id,version_no,content,snapshot_json,created_at) VALUES (?,?,?,?,NULL,?,?,?,?)`, v.ID, owner, c.ProjectID, c.ID, v.Version, v.Content, []byte(v.Snapshot), v.CreatedAt)
	return persistenceError(e)
}

func (s *MemoryStore) TouchProject(ctx context.Context, owner, id string) error {
	p, e := s.GetProject(ctx, owner, id)
	if e != nil {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	s.projects[memoryKey(owner, id)] = p
	return nil
}
func (s *MySQLStore) TouchProject(ctx context.Context, owner, id string) error {
	_, e := s.db.ExecContext(ctx, `UPDATE novel_projects SET updated_at=UTC_TIMESTAMP(6) WHERE owner_username=? AND project_id=? AND deleted_at IS NULL`, owner, id)
	return persistenceError(e)
}
