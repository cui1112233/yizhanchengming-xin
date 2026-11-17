package novelcreation

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// AtomicCreate starts an account-scoped transaction before a project exists.
// The receipt uniqueness lock is acquired first; LockImportProject follows it.
type ChapterImportStore interface {
	AtomicCreate(context.Context, string, func(Store) error) error
	ReserveSubmission(context.Context, string, string, string, string) (submissionReceipt, error)
	CompleteSubmission(context.Context, string, string, string, submissionReceipt) error
	CreateImportProject(context.Context, string, string) (Project, error)
	LockImportProject(context.Context, string, string) (Project, error)
	InsertImportedChapter(context.Context, string, Chapter, ChapterVersion) error
}

func (s *MySQLStore) AtomicCreate(ctx context.Context, owner string, fn func(Store) error) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return persistenceError(e)
	}
	defer tx.Rollback()
	if e = fn(&MySQLStore{db: tx}); e != nil {
		return e
	}
	return persistenceError(tx.Commit())
}
func (s *MySQLStore) ReserveSubmission(ctx context.Context, owner, kind, key, digest string) (out submissionReceipt, err error) {
	_, e := s.db.ExecContext(ctx, `INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,created_at) VALUES (?,?,?,?,UTC_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE idempotency_key=idempotency_key`, owner, kind, key, digest)
	if e != nil {
		return out, persistenceError(e)
	}
	var result []byte
	var project sql.NullString
	e = s.db.QueryRowContext(ctx, `SELECT fingerprint,project_id,result_json FROM novel_submission_receipts WHERE owner_username=? AND kind=? AND idempotency_key=? FOR UPDATE`, owner, kind, key).Scan(&out.Fingerprint, &project, &result)
	out.ProjectID = project.String
	out.Result = result
	return out, persistenceError(e)
}
func (s *MySQLStore) CompleteSubmission(ctx context.Context, owner, kind, key string, r submissionReceipt) error {
	result, e := s.db.ExecContext(ctx, `UPDATE novel_submission_receipts SET project_id=?,result_json=? WHERE owner_username=? AND kind=? AND idempotency_key=? AND fingerprint=? AND result_json IS NULL`, r.ProjectID, []byte(r.Result), owner, kind, key, r.Fingerprint)
	if e != nil {
		return persistenceError(e)
	}
	n, e := result.RowsAffected()
	if e != nil {
		return persistenceError(e)
	}
	if n != 1 {
		return conflict()
	}
	return nil
}
func (s *MySQLStore) CreateImportProject(ctx context.Context, owner, title string) (p Project, err error) {
	now, e := s.Clock(ctx)
	if e != nil {
		return p, e
	}
	p = Project{ID: novelID("project"), Title: title, Genre: "other", Status: "draft", CreatedAt: now, UpdatedAt: now}
	_, e = s.db.ExecContext(ctx, `INSERT INTO novel_projects (project_id,owner_username,title,genre,model_id,model_policy_json,target_chapter_words,status,created_at,updated_at) VALUES (?,?,?,?,'',?,0,?,?,?)`, p.ID, owner, p.Title, p.Genre, []byte(`{}`), p.Status, now, now)
	return p, persistenceError(e)
}
func (s *MySQLStore) LockImportProject(ctx context.Context, owner, id string) (Project, error) {
	p, e := scanProject(s.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM novel_projects WHERE owner_username=? AND project_id=? AND deleted_at IS NULL FOR UPDATE`, owner, id))
	if e != nil {
		return p, e
	}
	if projectReadOnly(p.Status) {
		return p, archivedError()
	}
	return p, nil
}
func (s *MySQLStore) InsertImportedChapter(ctx context.Context, owner string, c Chapter, v ChapterVersion) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO novel_chapters (chapter_id,owner_username,project_id,volume_id,number_no,title,content,word_count,audit_status,version_no,created_at,updated_at) VALUES (?,?,?,NULL,?,?,?,?,?,?,?,?)`, c.ID, owner, c.ProjectID, c.Number, c.Title, c.Content, c.WordCount, c.AuditStatus, c.Version, c.CreatedAt, c.UpdatedAt)
	if e != nil {
		return persistenceError(e)
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO novel_chapter_versions (version_id,owner_username,project_id,chapter_id,run_id,version_no,content,snapshot_json,created_at) VALUES (?,?,?,?,NULL,?,?,?,?)`, v.ID, owner, c.ProjectID, c.ID, v.Version, v.Content, []byte(v.Snapshot), v.CreatedAt)
	return persistenceError(e)
}
func (s *MemoryStore) ReserveSubmission(_ context.Context, owner, kind, key, digest string) (submissionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := memoryKey(owner, kind, key)
	r, ok := s.submissions[k]
	if !ok {
		r = submissionReceipt{Fingerprint: digest}
		s.submissions[k] = r
	}
	r.Result = cloneRawJSON(r.Result)
	return r, nil
}
func (s *MemoryStore) CompleteSubmission(_ context.Context, owner, kind, key string, r submissionReceipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := memoryKey(owner, kind, key)
	old, ok := s.submissions[k]
	if !ok || old.Fingerprint != r.Fingerprint || len(old.Result) > 0 {
		return conflict()
	}
	r.Result = cloneRawJSON(r.Result)
	s.submissions[k] = r
	return nil
}
func (s *MemoryStore) CreateImportProject(_ context.Context, owner, title string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	p := Project{ID: novelID("project"), Title: title, Genre: "other", Status: "draft", CreatedAt: now, UpdatedAt: now}
	s.projects[memoryKey(owner, p.ID)] = p
	return p, nil
}
func (s *MemoryStore) LockImportProject(ctx context.Context, owner, id string) (Project, error) {
	p, e := s.GetProject(ctx, owner, id)
	if e != nil {
		return p, e
	}
	if projectReadOnly(p.Status) {
		return p, archivedError()
	}
	return p, nil
}
func (s *MemoryStore) InsertImportedChapter(_ context.Context, owner string, c Chapter, v ChapterVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, b := range s.chapterBundles {
		if key == memoryKey(owner, c.ProjectID, c.ID) || (strings.HasPrefix(key, memoryKey(owner, c.ProjectID)+"\x00") && b.Chapter.Number == c.Number) {
			return conflict()
		}
	}
	s.chapterBundles[memoryKey(owner, c.ProjectID, c.ID)] = ChapterBundle{ProjectID: c.ProjectID, Chapter: c, ChapterVersion: v}
	s.chapterVersions[memoryKey(owner, c.ProjectID, c.ID, fmtVersion(v.Version))] = v
	return nil
}
