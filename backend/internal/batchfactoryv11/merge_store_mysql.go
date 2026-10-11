package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var _ MergeRepository = (*MySQLStore)(nil)

func (s *MySQLStore) FindMergeJob(ctx context.Context, owner, batchID, requestID string) (MergeJob, error) {
	return findMergeJob(ctx, s.db, owner, batchID, requestID)
}

func findMergeJob(ctx context.Context, q productionQueryer, owner, batchID, requestID string) (MergeJob, error) {
	var id string
	err := q.QueryRowContext(ctx, `SELECT id FROM batch_factory_v11_merge_jobs WHERE owner_username=? AND batch_id=? AND request_id=?`, owner, batchID, requestID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return MergeJob{}, ErrNotFound
	}
	if err != nil {
		return MergeJob{}, err
	}
	return loadMergeJob(ctx, q, owner, id)
}

func loadMergeJob(ctx context.Context, q productionQueryer, owner, id string) (MergeJob, error) {
	var value MergeJob
	var state string
	var outputURL, errorMessage, timingMode, progressPhase sql.NullString
	var providerTaskID sql.NullString
	err := q.QueryRowContext(ctx, `SELECT id,batch_id,COALESCE(book_id,''),COALESCE(opening_variant_index,0),request_id,timing_mode,speed,provider_task_id,status,progress_phase,progress_current,progress_total,output_url,error_message,created_at,updated_at FROM batch_factory_v11_merge_jobs WHERE id=? AND owner_username=?`, id, owner).Scan(
		&value.ID, &value.BatchID, &value.BookID, &value.OpeningVariantIndex, &value.RequestID, &timingMode, &value.Speed, &providerTaskID, &state, &progressPhase, &value.ProgressCurrent, &value.ProgressTotal, &outputURL, &errorMessage, &value.CreatedAt, &value.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return MergeJob{}, ErrNotFound
	}
	if err != nil {
		return MergeJob{}, err
	}
	value.Owner = owner
	value.Status = MergeState(state)
	value.TimingMode = timingMode.String
	value.ProgressPhase = progressPhase.String
	value.ProviderTaskID = providerTaskID.String
	value.OutputURL = outputURL.String
	value.ErrorMessage = errorMessage.String
	value.Sources = []MergeMedia{}
	rows, err := q.QueryContext(ctx, `SELECT video_id,ordinal,media_url,target_duration_seconds,requested_duration_seconds,actual_duration_seconds FROM batch_factory_v11_merge_sources WHERE job_id=? AND owner_username=? ORDER BY ordinal,video_id`, id, owner)
	if err != nil {
		return MergeJob{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var source MergeMedia
		if err := rows.Scan(&source.VideoID, &source.Order, &source.MediaURL, &source.TargetDurationSeconds, &source.RequestedDurationSeconds, &source.ActualDurationSeconds); err != nil {
			return MergeJob{}, err
		}
		source.URL = source.MediaURL
		value.Sources = append(value.Sources, source)
	}
	if err := rows.Err(); err != nil {
		return MergeJob{}, err
	}
	return value, nil
}

func (s *MySQLStore) CreateMergeJob(ctx context.Context, value MergeJob) (MergeJob, error) {
	if strings.TrimSpace(value.Owner) == "" || strings.TrimSpace(value.BatchID) == "" || strings.TrimSpace(value.RequestID) == "" || len(value.Sources) == 0 {
		return MergeJob{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MergeJob{}, err
	}
	defer tx.Rollback()
	if existing, findErr := findMergeJob(ctx, tx, value.Owner, value.BatchID, value.RequestID); findErr == nil {
		return existing, nil
	} else if !errors.Is(findErr, ErrNotFound) {
		return MergeJob{}, findErr
	}
	jobID, err := newID("merge")
	if err != nil {
		return MergeJob{}, err
	}
	now := time.Now().UTC()
	value.ID = jobID
	value.Status = normalizeMergeState(value.Status)
	value.CreatedAt, value.UpdatedAt = now, now
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_merge_jobs(id,owner_username,batch_id,book_id,opening_variant_index,request_id,timing_mode,speed,provider_task_id,status,progress_phase,progress_current,progress_total,output_url,error_message,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, value.BatchID, nullableString(value.BookID), value.OpeningVariantIndex, value.RequestID, nullableString(value.TimingMode), value.Speed, nullableString(value.ProviderTaskID), value.Status, nullableString(value.ProgressPhase), value.ProgressCurrent, value.ProgressTotal, nullableString(value.OutputURL), nullableString(value.ErrorMessage), now, now); err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			_ = tx.Rollback()
			return s.FindMergeJob(ctx, value.Owner, value.BatchID, value.RequestID)
		}
		return MergeJob{}, err
	}
	for ordinal, source := range value.Sources {
		if strings.TrimSpace(source.VideoID) == "" || strings.TrimSpace(source.MediaURL) == "" || source.Order != ordinal {
			return MergeJob{}, ErrInvalid
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_merge_sources(job_id,owner_username,video_id,ordinal,media_url,target_duration_seconds,requested_duration_seconds,actual_duration_seconds,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, source.VideoID, source.Order, source.MediaURL, source.TargetDurationSeconds, source.RequestedDurationSeconds, source.ActualDurationSeconds, now); err != nil {
			return MergeJob{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return MergeJob{}, err
	}
	return loadMergeJob(ctx, s.db, value.Owner, value.ID)
}

// CreateUploadedMergeJob records a manually uploaded final video that has no
// storyboard sources. Unlike CreateMergeJob it deliberately skips the
// sources-are-required contract because the uploaded MP4 is the whole deliverable.
func (s *MySQLStore) CreateUploadedMergeJob(ctx context.Context, value MergeJob) (MergeJob, error) {
	if strings.TrimSpace(value.Owner) == "" || strings.TrimSpace(value.BatchID) == "" || strings.TrimSpace(value.RequestID) == "" {
		return MergeJob{}, ErrInvalid
	}
	jobID, err := newID("merge")
	if err != nil {
		return MergeJob{}, err
	}
	now := time.Now().UTC()
	value.ID = jobID
	value.Status = normalizeMergeState(value.Status)
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v11_merge_jobs(id,owner_username,batch_id,book_id,request_id,timing_mode,speed,provider_task_id,status,progress_phase,progress_current,progress_total,output_url,error_message,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, value.BatchID, nullableString(value.BookID), value.RequestID, nullableString(value.TimingMode), value.Speed, nullableString(value.ProviderTaskID), value.Status, nullableString(value.ProgressPhase), value.ProgressCurrent, value.ProgressTotal, nullableString(value.OutputURL), nullableString(value.ErrorMessage), value.CreatedAt, value.UpdatedAt); err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			return s.FindMergeJob(ctx, value.Owner, value.BatchID, value.RequestID)
		}
		return MergeJob{}, err
	}
	return loadMergeJob(ctx, s.db, value.Owner, value.ID)
}

func (s *MySQLStore) UpdateMergeJob(ctx context.Context, owner, jobID string, value MergeJob) (MergeJob, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(jobID) == "" {
		return MergeJob{}, ErrInvalid
	}
	value.Status = normalizeMergeState(value.Status)
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MergeJob{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_merge_jobs SET opening_variant_index=?,timing_mode=?,speed=?,provider_task_id=?,status=?,progress_phase=?,progress_current=?,progress_total=?,output_url=?,error_message=?,updated_at=? WHERE id=? AND owner_username=?`, value.OpeningVariantIndex, nullableString(value.TimingMode), value.Speed, nullableString(value.ProviderTaskID), value.Status, nullableString(value.ProgressPhase), value.ProgressCurrent, value.ProgressTotal, nullableString(value.OutputURL), nullableString(value.ErrorMessage), now, jobID, owner)
	if err != nil {
		return MergeJob{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return MergeJob{}, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v11_merge_sources WHERE job_id=? AND owner_username=?`, jobID, owner); err != nil {
		return MergeJob{}, err
	}
	for ordinal, source := range value.Sources {
		if strings.TrimSpace(source.VideoID) == "" || strings.TrimSpace(source.MediaURL) == "" || source.Order != ordinal {
			return MergeJob{}, ErrInvalid
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_merge_sources(job_id,owner_username,video_id,ordinal,media_url,target_duration_seconds,requested_duration_seconds,actual_duration_seconds,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, jobID, owner, source.VideoID, source.Order, source.MediaURL, source.TargetDurationSeconds, source.RequestedDurationSeconds, source.ActualDurationSeconds, now); err != nil {
			return MergeJob{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return MergeJob{}, err
	}
	return loadMergeJob(ctx, s.db, owner, jobID)
}

func (s *MySQLStore) ListMergeJobs(ctx context.Context, owner, batchID string) ([]MergeJob, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM batch_factory_v11_merge_jobs WHERE owner_username=? AND batch_id=? ORDER BY created_at,id`, owner, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MergeJob{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		job, err := loadMergeJob(ctx, s.db, owner, id)
		if err != nil {
			return nil, err
		}
		out = append(out, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
