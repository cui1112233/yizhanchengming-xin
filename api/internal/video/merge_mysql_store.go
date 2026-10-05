package video

import (
	"context"
	"database/sql"
	"errors"
)

const mergeJobSelect = `SELECT id, batch_project_id, book_id, status, current_attempt, error_message, created_at, updated_at FROM video_merge_jobs`
const mergeAttemptSelect = `SELECT id, merge_job_id, attempt, status, aspect_ratio, speed, output_bucket, output_object_key, output_url, error_message, created_at, updated_at FROM video_merge_attempts`

func (s *MySQLStore) CreateMergeJob(ctx context.Context, job MergeJob) (MergeJob, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO video_merge_jobs
(batch_project_id, book_id, status, current_attempt, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`, job.BatchProjectID, job.BookID, job.Status, job.CurrentAttempt, job.ErrorMessage, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return MergeJob{}, err
	}
	job.ID, err = result.LastInsertId()
	if err != nil {
		return MergeJob{}, err
	}
	return s.GetMergeJob(ctx, job.ID)
}

func scanMergeJob(scan func(...any) error) (MergeJob, error) {
	var job MergeJob
	if err := scan(&job.ID, &job.BatchProjectID, &job.BookID, &job.Status, &job.CurrentAttempt, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MergeJob{}, ErrNotFound
		}
		return MergeJob{}, err
	}
	return job, nil
}

func (s *MySQLStore) GetMergeJob(ctx context.Context, id int64) (MergeJob, error) {
	return scanMergeJob(s.db.QueryRowContext(ctx, mergeJobSelect+` WHERE id=? LIMIT 1`, id).Scan)
}

func (s *MySQLStore) UpdateMergeJob(ctx context.Context, job MergeJob) error {
	result, err := s.db.ExecContext(ctx, `UPDATE video_merge_jobs SET status=?, current_attempt=?, error_message=?, updated_at=? WHERE id=?`, job.Status, job.CurrentAttempt, job.ErrorMessage, job.UpdatedAt, job.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MySQLStore) CreateMergeAttempt(ctx context.Context, attempt MergeAttempt) (MergeAttempt, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MergeAttempt{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO video_merge_attempts
(merge_job_id, attempt, status, aspect_ratio, speed, output_bucket, output_object_key, output_url, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, attempt.MergeJobID, attempt.Attempt, attempt.Status, attempt.AspectRatio, attempt.Speed, attempt.OutputBucket, attempt.OutputObjectKey, attempt.OutputURL, attempt.ErrorMessage, attempt.CreatedAt, attempt.UpdatedAt)
	if err != nil {
		return MergeAttempt{}, err
	}
	attempt.ID, err = result.LastInsertId()
	if err != nil {
		return MergeAttempt{}, err
	}
	for _, input := range attempt.Inputs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO video_merge_inputs
(merge_attempt_id, production_task_id, input_order, source_url, created_at)
VALUES (?, ?, ?, ?, ?)`, attempt.ID, input.ProductionTaskID, input.Order, input.URL, attempt.CreatedAt); err != nil {
			return MergeAttempt{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return MergeAttempt{}, err
	}
	return s.GetMergeAttempt(ctx, attempt.ID)
}

func scanMergeAttempt(scan func(...any) error) (MergeAttempt, error) {
	var attempt MergeAttempt
	if err := scan(&attempt.ID, &attempt.MergeJobID, &attempt.Attempt, &attempt.Status, &attempt.AspectRatio, &attempt.Speed, &attempt.OutputBucket, &attempt.OutputObjectKey, &attempt.OutputURL, &attempt.ErrorMessage, &attempt.CreatedAt, &attempt.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MergeAttempt{}, ErrNotFound
		}
		return MergeAttempt{}, err
	}
	return attempt, nil
}

func (s *MySQLStore) GetMergeAttempt(ctx context.Context, id int64) (MergeAttempt, error) {
	attempt, err := scanMergeAttempt(s.db.QueryRowContext(ctx, mergeAttemptSelect+` WHERE id=? LIMIT 1`, id).Scan)
	if err != nil {
		return MergeAttempt{}, err
	}
	attempt.Inputs, err = s.listMergeInputs(ctx, attempt.ID)
	return attempt, err
}

func (s *MySQLStore) UpdateMergeAttempt(ctx context.Context, attempt MergeAttempt) error {
	result, err := s.db.ExecContext(ctx, `UPDATE video_merge_attempts SET status=?, output_bucket=?, output_object_key=?, output_url=?, error_message=?, updated_at=? WHERE id=?`, attempt.Status, attempt.OutputBucket, attempt.OutputObjectKey, attempt.OutputURL, attempt.ErrorMessage, attempt.UpdatedAt, attempt.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *MySQLStore) ListMergeAttempts(ctx context.Context, jobID int64) ([]MergeAttempt, error) {
	rows, err := s.db.QueryContext(ctx, mergeAttemptSelect+` WHERE merge_job_id=? ORDER BY attempt ASC, id ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]MergeAttempt, 0)
	for rows.Next() {
		attempt, err := scanMergeAttempt(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, attempt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		inputs, err := s.listMergeInputs(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Inputs = inputs
	}
	return out, nil
}

func (s *MySQLStore) listMergeInputs(ctx context.Context, attemptID int64) ([]MergeInputAsset, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT production_task_id, input_order, source_url FROM video_merge_inputs WHERE merge_attempt_id=? ORDER BY input_order ASC, id ASC`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]MergeInputAsset, 0)
	for rows.Next() {
		var input MergeInputAsset
		if err := rows.Scan(&input.ProductionTaskID, &input.Order, &input.URL); err != nil {
			return nil, err
		}
		out = append(out, input)
	}
	return out, rows.Err()
}
