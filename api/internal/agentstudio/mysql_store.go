package agentstudio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }
func (s *MySQLStore) CreateProject(ctx context.Context, a Actor, in CreateProjectInput) (Project, error) {
	if s == nil || s.db == nil || a.UserID <= 0 {
		return Project{}, ErrInvalid
	}
	// The optional references are accepted only when the actor owns the BatchProject;
	// the book must belong to that referenced project.
	if in.BatchProjectID > 0 {
		var allowed bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM auth_batch_project_ownership WHERE batch_project_id=? AND (owner_user_id=? OR (team_id IS NOT NULL AND team_id=?)))`, in.BatchProjectID, a.UserID, a.TeamID).Scan(&allowed)
		if err != nil {
			return Project{}, err
		}
		if !allowed {
			return Project{}, ErrForbidden
		}
	}
	if in.BookID > 0 {
		if in.BatchProjectID <= 0 {
			return Project{}, ErrInvalid
		}
		var valid bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? AND b.id=?)`, in.BatchProjectID, in.BookID).Scan(&valid)
		if err != nil {
			return Project{}, err
		}
		if !valid {
			return Project{}, ErrInvalid
		}
	}
	r, err := s.db.ExecContext(ctx, `INSERT INTO agent_projects(owner_user_id,team_id,title,batch_project_id,book_id) VALUES(?,?,?,?,?)`, a.UserID, nullID(a.TeamID), in.Title, nullID(in.BatchProjectID), nullID(in.BookID))
	if err != nil {
		return Project{}, err
	}
	id, _ := r.LastInsertId()
	return s.GetProject(ctx, a, id)
}
func nullID(v int64) any {
	if v > 0 {
		return v
	}
	return nil
}
func (s *MySQLStore) GetProject(ctx context.Context, a Actor, id int64) (Project, error) {
	var p Project
	var team, batch, book sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT id,owner_user_id,team_id,title,batch_project_id,book_id,created_at,updated_at FROM agent_projects WHERE id=? AND (owner_user_id=? OR (team_id IS NOT NULL AND team_id=?))`, id, a.UserID, a.TeamID).Scan(&p.ID, &p.OwnerUserID, &team, &p.Title, &batch, &book, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("agent project: %w", err)
	}
	p.TeamID, p.BatchProjectID, p.BookID = team.Int64, batch.Int64, book.Int64
	return p, nil
}
func (s *MySQLStore) CreateMessage(ctx context.Context, m Message) (Message, error) {
	r, e := s.db.ExecContext(ctx, `INSERT INTO agent_messages(project_id,role,content,created_at)VALUES(?,?,?,?)`, m.ProjectID, m.Role, m.Content, m.CreatedAt)
	if e != nil {
		return Message{}, e
	}
	m.ID, _ = r.LastInsertId()
	return m, nil
}
func (s *MySQLStore) CreateExecution(ctx context.Context, e Execution) (Execution, error) {
	r, x := s.db.ExecContext(ctx, `INSERT INTO agent_executions(project_id,status,runtime_ref,error_code,error_message,created_at,updated_at)VALUES(?,?,?,?,?,?,?)`, e.ProjectID, e.Status, e.RuntimeRef, e.ErrorCode, e.ErrorMessage, e.CreatedAt, e.UpdatedAt)
	if x != nil {
		return Execution{}, x
	}
	e.ID, _ = r.LastInsertId()
	return e, nil
}
func (s *MySQLStore) UpdateExecution(ctx context.Context, e Execution) (Execution, error) {
	_, x := s.db.ExecContext(ctx, `UPDATE agent_executions SET status=?,runtime_ref=?,error_code=?,error_message=?,updated_at=? WHERE id=? AND project_id=?`, e.Status, e.RuntimeRef, e.ErrorCode, e.ErrorMessage, e.UpdatedAt, e.ID, e.ProjectID)
	return e, x
}
func (s *MySQLStore) ListProjects(ctx context.Context, a Actor) ([]Project, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,owner_user_id,COALESCE(team_id,0),title,COALESCE(batch_project_id,0),COALESCE(book_id,0),created_at,updated_at FROM agent_projects WHERE owner_user_id=? OR (team_id IS NOT NULL AND team_id=?) ORDER BY updated_at DESC`, a.UserID, a.TeamID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
		if e = rows.Scan(&p.ID, &p.OwnerUserID, &p.TeamID, &p.Title, &p.BatchProjectID, &p.BookID, &p.CreatedAt, &p.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *MySQLStore) ListMessages(ctx context.Context, a Actor, id int64) ([]Message, error) {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,project_id,role,content,created_at FROM agent_messages WHERE project_id=? ORDER BY id`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		e = rows.Scan(&m.ID, &m.ProjectID, &m.Role, &m.Content, &m.CreatedAt)
		if e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *MySQLStore) DeleteProject(ctx context.Context, a Actor, id int64) error {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return e
	}
	_, e := s.db.ExecContext(ctx, `DELETE FROM agent_projects WHERE id=?`, id)
	return e
}
func (s *MySQLStore) ListExecutions(ctx context.Context, a Actor, id int64) ([]Execution, error) {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,project_id,status,runtime_ref,error_code,error_message,created_at,updated_at FROM agent_executions WHERE project_id=? ORDER BY id DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Execution{}
	for rows.Next() {
		var v Execution
		if e = rows.Scan(&v.ID, &v.ProjectID, &v.Status, &v.RuntimeRef, &v.ErrorCode, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) CreateSkill(ctx context.Context, a Actor, in CreateSkillInput) (Skill, error) {
	if a.UserID <= 0 || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Body) == "" {
		return Skill{}, ErrInvalid
	}
	name, body := strings.TrimSpace(in.Name), strings.TrimSpace(in.Body)
	var version int
	if e := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0)+1 FROM agent_skill_versions WHERE owner_user_id=? AND name=?`, a.UserID, name).Scan(&version); e != nil {
		return Skill{}, e
	}
	r, e := s.db.ExecContext(ctx, `INSERT INTO agent_skill_versions(owner_user_id,name,version,body,enabled) VALUES(?,?,?,?,TRUE)`, a.UserID, name, version, body)
	if e != nil {
		return Skill{}, e
	}
	id, _ := r.LastInsertId()
	return Skill{ID: id, OwnerUserID: a.UserID, Name: name, Version: version, Body: body, Enabled: true}, nil
}
func (s *MySQLStore) ListSkills(ctx context.Context, a Actor) ([]Skill, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,owner_user_id,name,version,body,enabled,created_at FROM agent_skill_versions WHERE owner_user_id=? ORDER BY name,version DESC`, a.UserID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Skill{}
	for rows.Next() {
		var v Skill
		if e = rows.Scan(&v.ID, &v.OwnerUserID, &v.Name, &v.Version, &v.Body, &v.Enabled, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) GetSkills(ctx context.Context, a Actor, ids []int64) ([]Skill, error) {
	if len(ids) == 0 {
		return []Skill{}, nil
	}
	all, e := s.ListSkills(ctx, a)
	if e != nil {
		return nil, e
	}
	wanted := map[int64]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	out := []Skill{}
	for _, v := range all {
		if wanted[v.ID] && v.Enabled {
			out = append(out, v)
			delete(wanted, v.ID)
		}
	}
	if len(wanted) > 0 {
		return nil, ErrForbidden
	}
	return out, nil
}
func (s *MySQLStore) GetCanvas(ctx context.Context, a Actor, id int64) (Canvas, error) {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return Canvas{}, e
	}
	var c Canvas
	c.ProjectID = id
	e := s.db.QueryRowContext(ctx, `SELECT revision,document_json FROM agent_canvas_documents WHERE project_id=?`, id).Scan(&c.Revision, &c.Document)
	if errors.Is(e, sql.ErrNoRows) {
		return Canvas{ProjectID: id, Document: []byte(`{"nodes":[],"edges":[]}`)}, nil
	}
	return c, e
}
func (s *MySQLStore) SaveCanvas(ctx context.Context, a Actor, c Canvas) (Canvas, error) {
	if _, e := s.GetProject(ctx, a, c.ProjectID); e != nil {
		return Canvas{}, e
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return Canvas{}, e
	}
	defer tx.Rollback()
	r, e := tx.ExecContext(ctx, `UPDATE agent_canvas_documents SET document_json=?,revision=revision+1 WHERE project_id=? AND revision=?`, c.Document, c.ProjectID, c.Revision)
	if e != nil {
		return Canvas{}, e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		var x Canvas
		x.ProjectID = c.ProjectID
		e = tx.QueryRowContext(ctx, `SELECT revision,document_json FROM agent_canvas_documents WHERE project_id=?`, c.ProjectID).Scan(&x.Revision, &x.Document)
		if errors.Is(e, sql.ErrNoRows) {
			_, e = tx.ExecContext(ctx, `INSERT INTO agent_canvas_documents(project_id,revision,document_json) VALUES(?,?,?)`, c.ProjectID, 1, c.Document)
			if e == nil {
				x.Revision = 1
				x.Document = c.Document
				n = 1
			}
		}
		if e != nil {
			return Canvas{}, e
		}
		if n == 0 {
			return x, ErrConflict
		}
	}
	c.Revision++
	_, e = tx.ExecContext(ctx, `INSERT INTO agent_canvas_versions(project_id,revision,document_json)VALUES(?,?,?)`, c.ProjectID, c.Revision, c.Document)
	if e != nil {
		return Canvas{}, e
	}
	return c, tx.Commit()
}
func (s *MySQLStore) ListCanvasVersions(ctx context.Context, a Actor, id int64) ([]CanvasVersion, error) {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,project_id,revision,document_json,created_at FROM agent_canvas_versions WHERE project_id=? ORDER BY revision DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []CanvasVersion{}
	for rows.Next() {
		var v CanvasVersion
		if e = rows.Scan(&v.ID, &v.ProjectID, &v.Revision, &v.Document, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) RestoreCanvas(ctx context.Context, a Actor, id int64, revision int) (Canvas, error) {
	if _, e := s.GetProject(ctx, a, id); e != nil {
		return Canvas{}, e
	}
	var d []byte
	e := s.db.QueryRowContext(ctx, `SELECT document_json FROM agent_canvas_versions WHERE project_id=? AND revision=?`, id, revision).Scan(&d)
	if errors.Is(e, sql.ErrNoRows) {
		return Canvas{}, ErrNotFound
	}
	if e != nil {
		return Canvas{}, e
	}
	current, e := s.GetCanvas(ctx, a, id)
	if e != nil {
		return Canvas{}, e
	}
	current.Document = d
	return s.SaveCanvas(ctx, a, current)
}
func (s *MySQLStore) CreateAttachment(ctx context.Context, a Actor, v Attachment) (Attachment, error) {
	r, e := s.db.ExecContext(ctx, `INSERT INTO agent_attachments(project_id,owner_user_id,bucket,object_key,filename,content_type,byte_size,sha256) SELECT id,?,?,?,?,?,?,? FROM agent_projects WHERE id=? AND (owner_user_id=? OR (team_id IS NOT NULL AND team_id=?))`, v.ProjectID, a.UserID, v.Bucket, v.ObjectKey, v.Filename, v.ContentType, v.ByteSize, v.SHA256, v.ProjectID, a.UserID, a.TeamID)
	if e != nil {
		return Attachment{}, e
	}
	id, _ := r.LastInsertId()
	return s.GetAttachment(ctx, a, v.ProjectID, id)
}
func (s *MySQLStore) ListAttachments(ctx context.Context, a Actor, p int64) ([]Attachment, error) {
	if _, e := s.GetProject(ctx, a, p); e != nil {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT id,project_id,owner_user_id,bucket,object_key,filename,content_type,byte_size,sha256,created_at FROM agent_attachments WHERE project_id=? ORDER BY id DESC`, p)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Attachment{}
	for rows.Next() {
		var v Attachment
		if e = rows.Scan(&v.ID, &v.ProjectID, &v.OwnerUserID, &v.Bucket, &v.ObjectKey, &v.Filename, &v.ContentType, &v.ByteSize, &v.SHA256, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) GetAttachment(ctx context.Context, a Actor, p, id int64) (Attachment, error) {
	if _, e := s.GetProject(ctx, a, p); e != nil {
		return Attachment{}, e
	}
	var v Attachment
	e := s.db.QueryRowContext(ctx, `SELECT id,project_id,owner_user_id,bucket,object_key,filename,content_type,byte_size,sha256,created_at FROM agent_attachments WHERE project_id=? AND id=?`, p, id).Scan(&v.ID, &v.ProjectID, &v.OwnerUserID, &v.Bucket, &v.ObjectKey, &v.Filename, &v.ContentType, &v.ByteSize, &v.SHA256, &v.CreatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return Attachment{}, ErrNotFound
	}
	return v, e
}
func (s *MySQLStore) DeleteAttachment(ctx context.Context, a Actor, p, id int64) error {
	if _, e := s.GetAttachment(ctx, a, p, id); e != nil {
		return e
	}
	_, e := s.db.ExecContext(ctx, `DELETE FROM agent_attachments WHERE project_id=? AND id=?`, p, id)
	return e
}
