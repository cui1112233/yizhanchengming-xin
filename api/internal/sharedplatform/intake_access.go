package sharedplatform

import (
	"context"
	"database/sql"
	"fmt"
)

type MySQLIntakeAccessStore struct {
	db *sql.DB
}

func NewMySQLIntakeAccessStore(db *sql.DB) *MySQLIntakeAccessStore {
	return &MySQLIntakeAccessStore{db: db}
}

func (s *MySQLIntakeAccessStore) ClaimIntake(ctx context.Context, intakeID, ownerUserID, teamID int64) error {
	if s == nil || s.db == nil || intakeID <= 0 || ownerUserID <= 0 {
		return fmt.Errorf("shared platform: intake ownership unavailable")
	}
	var team any
	if teamID > 0 {
		team = teamID
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE intake_id = VALUES(intake_id)`, intakeID, ownerUserID, team)
	if err != nil {
		return fmt.Errorf("shared platform: claim intake ownership: %w", err)
	}
	return nil
}

func (s *MySQLIntakeAccessStore) CanAccessIntake(ctx context.Context, intakeID, userID, teamID int64, elevated bool) (bool, error) {
	if s == nil || s.db == nil || intakeID <= 0 || userID <= 0 {
		return false, fmt.Errorf("shared platform: intake access unavailable")
	}
	var allowed bool
	if elevated {
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM intakes WHERE id = ?)`, intakeID).Scan(&allowed)
		return allowed, err
	}
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(
        SELECT 1 FROM auth_intake_ownership
        WHERE intake_id = ?
          AND (owner_user_id = ? OR (team_id IS NOT NULL AND team_id = ?))
    )`, intakeID, userID, teamID).Scan(&allowed)
	return allowed, err
}
