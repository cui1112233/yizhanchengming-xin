package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const selectByIdempotencySQL = `SELECT id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at FROM magic_cut_projects WHERE owner_username = ? AND idempotency_key = ? LIMIT 1`
const insertProjectSQL = `INSERT INTO magic_cut_projects (id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
const listProjectsSQL = `SELECT id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at FROM magic_cut_projects WHERE owner_username = ? ORDER BY updated_at DESC`

type Store struct {
	db    *sql.DB
	now   func() time.Time
	newID func() string
}

func NewStore(db *sql.DB, now func() time.Time, newID func() string) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{db: db, now: now, newID: newID}
}

func (s *Store) Create(ctx context.Context, owner string, input CreateInput) (Project, error) {
	owner = strings.TrimSpace(owner)
	input.Name = strings.TrimSpace(input.Name)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if owner == "" || input.Name == "" || input.IdempotencyKey == "" {
		return Project{}, errors.New("owner, name and idempotencyKey are required")
	}
	if input.Width <= 0 || input.Height <= 0 || input.FrameRate <= 0 {
		return Project{}, errors.New("width, height and frameRate must be positive")
	}
	if s == nil || s.db == nil || s.newID == nil {
		return Project{}, errors.New("project store is not configured")
	}

	existing, err := s.byIdempotency(ctx, owner, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Project{}, fmt.Errorf("find idempotent project: %w", err)
	}

	now := s.now().UTC()
	created := Project{
		ID:             s.newID(),
		Owner:          owner,
		Name:           input.Name,
		Width:          input.Width,
		Height:         input.Height,
		FrameRate:      input.FrameRate,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	_, err = s.db.ExecContext(ctx, insertProjectSQL, created.ID, created.Owner, created.Name, created.Width, created.Height, created.FrameRate, created.IdempotencyKey, created.CreatedAt, created.UpdatedAt)
	if err != nil {
		if raced, lookupErr := s.byIdempotency(ctx, owner, input.IdempotencyKey); lookupErr == nil {
			return raced, nil
		}
		return Project{}, fmt.Errorf("insert project: %w", err)
	}
	return created, nil
}

func (s *Store) List(ctx context.Context, owner string) ([]Project, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("project store is not configured")
	}
	rows, err := s.db.QueryContext(ctx, listProjectsSQL, strings.TrimSpace(owner))
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		project, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (s *Store) byIdempotency(ctx context.Context, owner, key string) (Project, error) {
	return scanProject(s.db.QueryRowContext(ctx, selectByIdempotencySQL, owner, key))
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(row scanner) (Project, error) {
	var project Project
	err := row.Scan(&project.ID, &project.Owner, &project.Name, &project.Width, &project.Height, &project.FrameRate, &project.IdempotencyKey, &project.CreatedAt, &project.UpdatedAt)
	return project, err
}
