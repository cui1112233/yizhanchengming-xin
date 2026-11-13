package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type MySQLAuthoringRepository struct{ db *sql.DB }

func (s *MySQLStore) AuthoringRepository() AuthoringRepository {
	return &MySQLAuthoringRepository{db: s.db}
}
func (r *MySQLAuthoringRepository) Load(ctx context.Context, o, p string) (AuthoringState, error) {
	var raw []byte
	var v int64
	e := r.db.QueryRowContext(ctx, `SELECT body,revision FROM screenplay_authoring WHERE owner_username=? AND project_id=?`, o, p).Scan(&raw, &v)
	if errors.Is(e, sql.ErrNoRows) {
		return AuthoringState{}, ErrNotFound
	}
	if e != nil {
		return AuthoringState{}, e
	}
	var st AuthoringState
	e = json.Unmarshal(raw, &st)
	st.Revision = v
	return st, e
}
func (r *MySQLAuthoringRepository) Save(ctx context.Context, st AuthoringState, v int64) (AuthoringState, error) {
	st.Revision = v + 1
	b, e := json.Marshal(st)
	if e != nil {
		return st, e
	}
	var result sql.Result
	if v == 0 {
		result, e = r.db.ExecContext(ctx, `INSERT IGNORE INTO screenplay_authoring(project_id,owner_username,revision,body) VALUES(?,?,?,?)`, st.ProjectID, st.Owner, st.Revision, b)
	} else {
		result, e = r.db.ExecContext(ctx, `UPDATE screenplay_authoring SET body=?,revision=? WHERE project_id=? AND owner_username=? AND revision=?`, b, st.Revision, st.ProjectID, st.Owner, v)
	}
	if e != nil {
		return AuthoringState{}, e
	}
	n, e := result.RowsAffected()
	if e == nil && n != 1 {
		return AuthoringState{}, ErrRevisionConflict
	}
	return st, e
}
func (r *MySQLAuthoringRepository) List(ctx context.Context) ([]AuthoringState, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT body,revision FROM screenplay_authoring`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []AuthoringState{}
	for rows.Next() {
		var b []byte
		var st AuthoringState
		var v int64
		if e = rows.Scan(&b, &v); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &st); e != nil {
			return nil, e
		}
		st.Revision = v
		out = append(out, st)
	}
	return out, rows.Err()
}
