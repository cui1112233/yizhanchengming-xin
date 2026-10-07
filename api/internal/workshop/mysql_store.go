package workshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) Load(ctx context.Context, intakeID int64) (json.RawMessage, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(workshop_settings_json, JSON_OBJECT()) FROM intakes WHERE id=?`, intakeID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}
func (s *MySQLStore) Save(ctx context.Context, intakeID int64, settings json.RawMessage) (json.RawMessage, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE intakes SET workshop_settings_json=? WHERE id=?`, settings, intakeID)
	if err != nil {
		return nil, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, ErrNotFound
	}
	return settings, nil
}
