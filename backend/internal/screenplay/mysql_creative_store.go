package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *MySQLStore) ListUserStyles(ctx context.Context, owner string) ([]UserStyle, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,name,prompt,images,revision,created_at,updated_at FROM screenplay_user_styles WHERE owner_username=? ORDER BY updated_at DESC,id`, owner)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []UserStyle{}
	for rows.Next() {
		var v UserStyle
		var b []byte
		v.OwnerUsername = owner
		if e = rows.Scan(&v.ID, &v.Name, &v.Prompt, &b, &v.Revision, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &v.Images); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *MySQLStore) GetUserStyle(ctx context.Context, owner, id string) (UserStyle, error) {
	v := UserStyle{OwnerUsername: owner}
	var b []byte
	e := s.db.QueryRowContext(ctx, `SELECT id,name,prompt,images,revision,created_at,updated_at FROM screenplay_user_styles WHERE owner_username=? AND id=?`, owner, id).Scan(&v.ID, &v.Name, &v.Prompt, &b, &v.Revision, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	if e != nil {
		return v, e
	}
	e = json.Unmarshal(b, &v.Images)
	return v, e
}
func (s *MySQLStore) SaveUserStyle(ctx context.Context, v UserStyle, expected *int64) (UserStyle, error) {
	if v.ID == "" || v.OwnerUsername == "" || v.Name == "" || v.Prompt == "" {
		return v, ErrInvalidArgument
	}
	if v.Images == nil {
		v.Images = []string{}
	}
	b, e := json.Marshal(v.Images)
	if e != nil {
		return v, e
	}
	now := time.Now().UTC()
	if expected == nil {
		_, e = s.db.ExecContext(ctx, `INSERT INTO screenplay_user_styles(owner_username,id,name,prompt,images,revision,created_at,updated_at) VALUES(?,?,?,?,?,0,?,?)`, v.OwnerUsername, v.ID, v.Name, v.Prompt, string(b), now, now)
	} else {
		var res sql.Result
		res, e = s.db.ExecContext(ctx, `UPDATE screenplay_user_styles SET name=?,prompt=?,images=?,revision=revision+1,updated_at=? WHERE owner_username=? AND id=? AND revision=?`, v.Name, v.Prompt, string(b), now, v.OwnerUsername, v.ID, *expected)
		if e == nil {
			n, _ := res.RowsAffected()
			if n != 1 {
				return v, ErrRevisionConflict
			}
		}
	}
	if e != nil {
		return v, e
	}
	return s.GetUserStyle(ctx, v.OwnerUsername, v.ID)
}
func (s *MySQLStore) DeleteUserStyle(ctx context.Context, owner, id string, expected int64) error {
	res, e := s.db.ExecContext(ctx, `DELETE FROM screenplay_user_styles WHERE owner_username=? AND id=? AND revision=?`, owner, id, expected)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return ErrRevisionConflict
	}
	return nil
}
func (s *MySQLStore) loadCreativeConfig(ctx context.Context, p *Project) error {
	var b []byte
	e := s.db.QueryRowContext(ctx, `SELECT config FROM screenplay_project_creative_configs WHERE project_id=? AND owner_username=?`, p.ID, p.OwnerUsername).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	p.CreativeConfig, e = decodeCreativeConfig(b)
	return e
}
func (s *MySQLStore) SaveCreativeConfig(ctx context.Context, owner, id string, c CreativeConfig, expected int64) (Project, error) {
	b, e := encodeCreativeConfig(&c)
	if e != nil {
		return Project{}, e
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return Project{}, e
	}
	defer tx.Rollback()
	var revision int64
	e = tx.QueryRowContext(ctx, `SELECT revision FROM screenplay_projects WHERE id=? AND owner_username=? AND deleted_at IS NULL FOR UPDATE`, id, owner).Scan(&revision)
	if errors.Is(e, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if e != nil {
		return Project{}, e
	}
	if revision != expected {
		return Project{}, ErrRevisionConflict
	}
	p := Project{}
	applyCreativeConfig(&p, c)
	now := time.Now().UTC()
	// The snapshot and compatibility fields are changed under the same project lock.
	res, e := tx.ExecContext(ctx, `UPDATE screenplay_projects SET art_style=?,style_desc=?,selected_style=?,prompt_id=IF(?='',prompt_id,?),aspect_ratio=IF(?='',aspect_ratio,?),revision=revision+1,updated_at=? WHERE id=? AND owner_username=? AND revision=? AND deleted_at IS NULL`, p.ArtStyle, p.StyleDesc, p.SelectedStyle, c.ScriptPromptID, c.ScriptPromptID, c.AspectRatio, c.AspectRatio, now, id, owner, expected)
	if e != nil {
		return Project{}, e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return Project{}, ErrRevisionConflict
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO screenplay_project_creative_configs(project_id,owner_username,config,updated_at) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE config=VALUES(config),updated_at=VALUES(updated_at)`, id, owner, string(b), now)
	if e != nil {
		return Project{}, fmt.Errorf("保存创作配置: %w", e)
	}
	if e = tx.Commit(); e != nil {
		return Project{}, e
	}
	return s.GetProject(ctx, owner, id)
}
