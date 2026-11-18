package novelcreation

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type LongImitationStore interface {
	CreateLongImitationProject(context.Context, string, string, string, int) (Project, error)
	GetSubmissionReceipt(context.Context, string, string, string) (submissionReceipt, bool, error)
}

func (s *MemoryStore) CreateLongImitationProject(ctx context.Context, owner, title, model string, target int) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	p := Project{ID: novelID("project"), Title: title, Genre: "other", ModelID: model, TargetChapterWords: target, Status: "draft", CreatedAt: now, UpdatedAt: now}
	s.projects[memoryKey(owner, p.ID)] = p
	return p, nil
}
func (s *MySQLStore) CreateLongImitationProject(ctx context.Context, owner, title, model string, target int) (Project, error) {
	now := time.Now().UTC()
	p := Project{ID: novelID("project"), Title: title, Genre: "other", ModelID: model, TargetChapterWords: target, Status: "draft", CreatedAt: now, UpdatedAt: now}
	_, e := s.db.ExecContext(ctx, `INSERT INTO novel_projects (project_id,owner_username,title,genre,model_id,model_policy_json,target_chapter_words,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`, p.ID, owner, title, "other", model, []byte(`{"creationMode":"long_imitation"}`), target, "draft", now, now)
	return p, persistenceError(e)
}
func (s *MemoryStore) GetSubmissionReceipt(_ context.Context, owner, kind, key string) (submissionReceipt, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.submissions[memoryKey(owner, kind, key)]
	r.Result = cloneRawJSON(r.Result)
	return r, ok, nil
}
func (s *MySQLStore) GetSubmissionReceipt(ctx context.Context, owner, kind, key string) (out submissionReceipt, found bool, err error) {
	var project sql.NullString
	e := s.db.QueryRowContext(ctx, `SELECT fingerprint,project_id,result_json FROM novel_submission_receipts WHERE owner_username=? AND kind=? AND idempotency_key=?`, owner, kind, key).Scan(&out.Fingerprint, &project, &out.Result)
	if errors.Is(e, sql.ErrNoRows) {
		return out, false, nil
	}
	out.ProjectID = project.String
	return out, e == nil, persistenceError(e)
}
