package video

import (
	"context"
	"database/sql"
	"errors"
)

func (s *MySQLStore) ProjectIDForProductionTask(ctx context.Context, taskID int64) (int64, error) {
	var projectID int64
	err := s.db.QueryRowContext(ctx, `
SELECT j.batch_project_id
FROM video_production_tasks t
JOIN video_production_jobs j ON j.id = t.production_job_id
WHERE t.id = ?
LIMIT 1`, taskID).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return projectID, err
}

func (s *MySQLStore) ProjectIDForMergeJob(ctx context.Context, jobID int64) (int64, error) {
	var projectID int64
	err := s.db.QueryRowContext(ctx, `SELECT batch_project_id FROM video_merge_jobs WHERE id = ? LIMIT 1`, jobID).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return projectID, err
}

func (s *MySQLStore) ProjectIDForMergeAttempt(ctx context.Context, attemptID int64) (int64, error) {
	var projectID int64
	err := s.db.QueryRowContext(ctx, `
SELECT j.batch_project_id
FROM video_merge_attempts a
JOIN video_merge_jobs j ON j.id = a.merge_job_id
WHERE a.id = ?
LIMIT 1`, attemptID).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return projectID, err
}
