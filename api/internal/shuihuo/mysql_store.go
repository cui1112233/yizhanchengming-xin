package shuihuo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) CreateProject(ctx context.Context, project Project) (Project, error) {
	if s == nil || s.db == nil {
		return Project{}, ErrInvalid
	}
	var team any
	if project.TeamID > 0 {
		team = project.TeamID
	}
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO shuihuo_projects (owner_user_id, team_id, name, production_mode, source_text, segmentation_status) VALUES (?, ?, ?, ?, ?, ?)`,
		project.OwnerUserID, team, project.Name, project.ProductionMode, project.SourceText, project.SegmentationStatus,
	)
	if err != nil {
		return Project{}, fmt.Errorf("create shuihuo project: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Project{}, fmt.Errorf("read shuihuo project id: %w", err)
	}
	return s.GetProject(ctx, id)
}

func (s *MySQLStore) ListProjects(ctx context.Context, actor Actor) ([]Project, error) {
	if s == nil || s.db == nil {
		return nil, ErrInvalid
	}
	query := `SELECT id, owner_user_id, team_id, name, production_mode, source_text, segmentation_status, created_at, updated_at FROM shuihuo_projects`
	args := []any{}
	if !actor.BypassOwnership {
		if actor.TeamID > 0 {
			query += ` WHERE owner_user_id = ? OR team_id = ?`
			args = append(args, actor.UserID, actor.TeamID)
		} else {
			query += ` WHERE owner_user_id = ?`
			args = append(args, actor.UserID)
		}
	}
	query += ` ORDER BY updated_at DESC, id DESC LIMIT 200`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list shuihuo projects: %w", err)
	}
	defer rows.Close()
	out := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan shuihuo project: %w", err)
		}
		out = append(out, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shuihuo projects: %w", err)
	}
	return out, nil
}

func (s *MySQLStore) GetProject(ctx context.Context, id int64) (Project, error) {
	if s == nil || s.db == nil {
		return Project{}, ErrInvalid
	}
	project, err := scanProject(s.db.QueryRowContext(ctx,
		`SELECT id, owner_user_id, team_id, name, production_mode, source_text, segmentation_status, created_at, updated_at FROM shuihuo_projects WHERE id = ?`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("get shuihuo project: %w", err)
	}
	return project, nil
}

func (s *MySQLStore) SaveSourceAndReset(ctx context.Context, id int64, source string) (Project, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, fmt.Errorf("begin source reset: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE shuihuo_projects SET source_text = ?, segmentation_status = ?, updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, source, SegmentationDraft, id)
	if err != nil {
		return Project{}, fmt.Errorf("save shuihuo source: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Project{}, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM shuihuo_segments WHERE project_id = ?`, id); err != nil {
		return Project{}, fmt.Errorf("clear stale shuihuo segments: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Project{}, fmt.Errorf("commit source reset: %w", err)
	}
	return s.GetProject(ctx, id)
}

func (s *MySQLStore) DeleteProject(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM shuihuo_projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete shuihuo project: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MySQLStore) ReplaceSegments(ctx context.Context, projectID int64, candidates []Candidate) ([]Segment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin segmentation confirm: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE shuihuo_projects SET segmentation_status = ?, updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, SegmentationConfirmed, projectID)
	if err != nil {
		return nil, fmt.Errorf("confirm shuihuo segmentation: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM shuihuo_segments WHERE project_id = ?`, projectID); err != nil {
		return nil, fmt.Errorf("replace shuihuo segments: %w", err)
	}
	for i, candidate := range candidates {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO shuihuo_segments (project_id, position, source_text, subtitle_text, speaker) VALUES (?, ?, ?, ?, ?)`,
			projectID, i+1, candidate.Text, candidate.Text, normalizeSpeaker(candidate.Speaker),
		); err != nil {
			return nil, fmt.Errorf("insert shuihuo segment: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit segmentation confirm: %w", err)
	}
	return s.ListSegments(ctx, projectID)
}

func (s *MySQLStore) ListSegments(ctx context.Context, projectID int64) ([]Segment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, position, source_text, subtitle_text, speaker, created_at, updated_at FROM shuihuo_segments WHERE project_id = ? ORDER BY position ASC, id ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list shuihuo segments: %w", err)
	}
	defer rows.Close()
	out := make([]Segment, 0)
	for rows.Next() {
		segment, err := scanSegment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan shuihuo segment: %w", err)
		}
		out = append(out, segment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shuihuo segments: %w", err)
	}
	return out, nil
}

func (s *MySQLStore) GetSegment(ctx context.Context, segmentID int64) (Segment, error) {
	segment, err := scanSegment(s.db.QueryRowContext(ctx, `SELECT id, project_id, position, source_text, subtitle_text, speaker, created_at, updated_at FROM shuihuo_segments WHERE id = ?`, segmentID))
	if errors.Is(err, sql.ErrNoRows) {
		return Segment{}, ErrNotFound
	}
	if err != nil {
		return Segment{}, fmt.Errorf("get shuihuo segment: %w", err)
	}
	return segment, nil
}

func (s *MySQLStore) CreateSegment(ctx context.Context, projectID int64, input SegmentInput) (Segment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Segment{}, fmt.Errorf("begin segment create: %w", err)
	}
	defer tx.Rollback()
	var position int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM shuihuo_segments WHERE project_id = ?`, projectID).Scan(&position); err != nil {
		return Segment{}, fmt.Errorf("read next shuihuo segment position: %w", err)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO shuihuo_segments (project_id, position, source_text, subtitle_text, speaker) VALUES (?, ?, ?, ?, ?)`, projectID, position, input.SourceText, input.SubtitleText, normalizeSpeaker(input.Speaker))
	if err != nil {
		return Segment{}, fmt.Errorf("create shuihuo segment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Segment{}, fmt.Errorf("read shuihuo segment id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_projects SET updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, projectID); err != nil {
		return Segment{}, fmt.Errorf("touch shuihuo project: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Segment{}, fmt.Errorf("commit segment create: %w", err)
	}
	return s.GetSegment(ctx, id)
}

func (s *MySQLStore) UpdateSegment(ctx context.Context, segmentID int64, input SegmentInput) (Segment, error) {
	current, err := s.GetSegment(ctx, segmentID)
	if err != nil {
		return Segment{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE shuihuo_segments SET source_text = ?, subtitle_text = ?, speaker = ?, updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, input.SourceText, input.SubtitleText, normalizeSpeaker(input.Speaker), segmentID)
	if err != nil {
		return Segment{}, fmt.Errorf("update shuihuo segment: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Segment{}, ErrNotFound
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE shuihuo_projects SET updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, current.ProjectID)
	return s.GetSegment(ctx, segmentID)
}

func (s *MySQLStore) DeleteSegment(ctx context.Context, segmentID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin segment delete: %w", err)
	}
	defer tx.Rollback()
	var projectID int64
	var position int
	if err := tx.QueryRowContext(ctx, `SELECT project_id, position FROM shuihuo_segments WHERE id = ? FOR UPDATE`, segmentID).Scan(&projectID, &position); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("read shuihuo segment for delete: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM shuihuo_segments WHERE id = ?`, segmentID); err != nil {
		return fmt.Errorf("delete shuihuo segment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_segments SET position = position - 1 WHERE project_id = ? AND position > ? ORDER BY position ASC`, projectID, position); err != nil {
		return fmt.Errorf("compact shuihuo segment positions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_projects SET updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, projectID); err != nil {
		return fmt.Errorf("touch shuihuo project: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit segment delete: %w", err)
	}
	return nil
}

func (s *MySQLStore) ReorderSegments(ctx context.Context, projectID int64, ids []int64) ([]Segment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin segment reorder: %w", err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id FROM shuihuo_segments WHERE project_id = ? ORDER BY position ASC FOR UPDATE`, projectID)
	if err != nil {
		return nil, fmt.Errorf("read shuihuo segment order: %w", err)
	}
	existing := make(map[int64]struct{})
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan shuihuo segment id: %w", err)
		}
		existing[id] = struct{}{}
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close shuihuo segment rows: %w", err)
	}
	if len(existing) != len(ids) || len(ids) == 0 {
		return nil, ErrInvalid
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := existing[id]; !ok {
			return nil, ErrInvalid
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, ErrInvalid
		}
		seen[id] = struct{}{}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_segments SET position = position + 1000000 WHERE project_id = ?`, projectID); err != nil {
		return nil, fmt.Errorf("stage shuihuo reorder: %w", err)
	}
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_segments SET position = ? WHERE id = ? AND project_id = ?`, i+1, id, projectID); err != nil {
			return nil, fmt.Errorf("apply shuihuo reorder: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shuihuo_projects SET updated_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, projectID); err != nil {
		return nil, fmt.Errorf("touch shuihuo project: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit shuihuo reorder: %w", err)
	}
	return s.ListSegments(ctx, projectID)
}

type rowScanner interface{ Scan(...any) error }

func scanProject(row rowScanner) (Project, error) {
	var value Project
	var team sql.NullInt64
	if err := row.Scan(&value.ID, &value.OwnerUserID, &team, &value.Name, &value.ProductionMode, &value.SourceText, &value.SegmentationStatus, &value.CreatedAt, &value.UpdatedAt); err != nil {
		return Project{}, err
	}
	if team.Valid {
		value.TeamID = team.Int64
	}
	return value, nil
}

func scanSegment(row rowScanner) (Segment, error) {
	var value Segment
	if err := row.Scan(&value.ID, &value.ProjectID, &value.Position, &value.SourceText, &value.SubtitleText, &value.Speaker, &value.CreatedAt, &value.UpdatedAt); err != nil {
		return Segment{}, err
	}
	return value, nil
}
