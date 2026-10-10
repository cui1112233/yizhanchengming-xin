package localexecutor

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const jobSelect = `SELECT id, owner_username, source_task_id, platform, payload_json, state,
 cancel_requested, lease_executor_id, lease_token_hash, lease_generation, lease_expires_at,
 accepted_at, accepted_account_id, submission_id, artifact_id, error_code, error_message,
 created_at, updated_at FROM local_executor_jobs`

func (s *MySQLStore) CreateJob(ctx context.Context, record JobRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO local_executor_jobs
(id, owner_username, source_task_id, platform, payload_json, state, cancel_requested,
 lease_executor_id, lease_token_hash, lease_generation, lease_expires_at, accepted_at,
 accepted_account_id, submission_id, artifact_id, error_code, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, FALSE, NULL, NULL, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, ?, ?)`,
		// Pass JSON as a string so MySQL treats it as JSON text rather than a
		// binary value. Passing []byte makes MySQL 8 reject the parameter with
		// error 3144 ("Cannot create a JSON value ... CHARACTER SET 'binary'").
		record.ID, record.OwnerUsername, record.SourceTaskID, record.Platform, string(record.Payload), record.State,
		record.CreatedAt, record.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *MySQLStore) JobForOwner(ctx context.Context, owner, id string) (JobRecord, error) {
	record, err := scanJob(s.db.QueryRowContext(ctx, jobSelect+` WHERE id = ? AND owner_username = ?`, id, owner).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	return record, err
}

func (s *MySQLStore) CancelJob(ctx context.Context, owner, id string, now time.Time) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := scanJob(tx.QueryRowContext(ctx, jobSelect+` WHERE id = ? AND owner_username = ? FOR UPDATE`, id, owner).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if record.State != JobSucceeded && record.State != JobFailed && record.State != JobCancelled {
		record.CancelRequested = true
		record.State = JobCancelled
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET cancel_requested = TRUE, state = ?, updated_at = ? WHERE id = ?`, JobCancelled, now, id); err != nil {
			return JobRecord{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) ClaimJob(ctx context.Context, executor ExecutorRecord, leaseHash SecretHash, expires, now time.Time) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()

	record, err := scanJob(tx.QueryRowContext(ctx, jobSelect+` WHERE owner_username = ? AND platform = ?
 AND cancel_requested = FALSE AND accepted_at IS NULL
 AND (state = ? OR (lease_expires_at IS NOT NULL AND lease_expires_at <= ? AND state NOT IN (?, ?, ?)))
 ORDER BY created_at ASC, id ASC LIMIT 1 FOR UPDATE SKIP LOCKED`,
		executor.OwnerUsername, executor.Platform, JobQueued, now, JobSucceeded, JobFailed, JobCancelled).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrNoClaimableJob
	}
	if err != nil {
		return JobRecord{}, err
	}

	record.State = JobLeased
	record.LeaseExecutorID = executor.ID
	record.LeaseTokenHash = leaseHash
	record.LeaseGeneration++
	record.LeaseExpiresAt = &expires
	record.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, lease_executor_id = ?, lease_token_hash = ?,
 lease_generation = ?, lease_expires_at = ?, error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ?`,
		JobLeased, executor.ID, leaseHash[:], record.LeaseGeneration, expires, now, record.ID); err != nil {
		return JobRecord{}, err
	}
	if err := appendJobEvent(ctx, tx, record.ID, executor.ID, "claimed", JobLeased, "", now); err != nil {
		return JobRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) RenewJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, expires, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		record.LeaseExpiresAt = &expires
		record.UpdatedAt = now
		_, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET lease_expires_at = ?, updated_at = ? WHERE id = ?`, expires, now, record.ID)
		return err
	})
}

func (s *MySQLStore) SetJobState(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, next JobState, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if err := validateTransition(record.State, next, isAccepted(*record)); err != nil {
			return err
		}
		record.State = next
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, updated_at = ? WHERE id = ?`, next, now, record.ID); err != nil {
			return err
		}
		return appendJobEvent(ctx, tx, record.ID, executorID, "progress", next, "", now)
	})
}

func (s *MySQLStore) AcceptJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, accountID, submissionID string, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if isAccepted(*record) {
			if record.AcceptedAccountID == accountID && record.SubmissionID == submissionID {
				return nil
			}
			return ErrJobConflict
		}
		if record.State != JobSubmitting && record.State != JobAcceptanceUnknown {
			return ErrInvalidJobState
		}
		record.State = JobAccepted
		record.AcceptedAt = &now
		record.AcceptedAccountID = accountID
		record.SubmissionID = submissionID
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, accepted_at = ?, accepted_account_id = ?, submission_id = ?, updated_at = ? WHERE id = ?`,
			JobAccepted, now, accountID, submissionID, now, record.ID); err != nil {
			return err
		}
		return appendJobEvent(ctx, tx, record.ID, executorID, "accepted", JobAccepted, submissionID, now)
	})
}

func (s *MySQLStore) ReleaseJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, reason string, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if isAccepted(*record) {
			return ErrAcceptedJobPinned
		}
		record.State = JobQueued
		record.LeaseExecutorID = ""
		record.LeaseTokenHash = SecretHash{}
		record.LeaseExpiresAt = nil
		record.ErrorMessage = bounded(reason, 512)
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, lease_executor_id = NULL,
 lease_token_hash = NULL, lease_expires_at = NULL, error_message = ?, updated_at = ? WHERE id = ?`,
			JobQueued, record.ErrorMessage, now, record.ID); err != nil {
			return err
		}
		return appendJobEvent(ctx, tx, record.ID, executorID, "released", JobQueued, record.ErrorMessage, now)
	})
}

func (s *MySQLStore) FailJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, code, message string, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		record.State = JobFailed
		record.ErrorCode = bounded(code, 96)
		record.ErrorMessage = bounded(message, 512)
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, error_code = ?, error_message = ?, updated_at = ? WHERE id = ?`,
			JobFailed, record.ErrorCode, record.ErrorMessage, now, record.ID); err != nil {
			return err
		}
		return appendJobEvent(ctx, tx, record.ID, executorID, "failed", JobFailed, record.ErrorCode, now)
	})
}

func (s *MySQLStore) CompleteJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, artifactID string, now time.Time) (JobRecord, error) {
	return s.mutateJob(ctx, executorID, jobID, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if !isAccepted(*record) {
			return ErrInvalidJobState
		}
		record.State = JobSucceeded
		record.ArtifactID = artifactID
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE local_executor_jobs SET state = ?, artifact_id = ?, updated_at = ? WHERE id = ?`,
			JobSucceeded, artifactID, now, record.ID); err != nil {
			return err
		}
		return appendJobEvent(ctx, tx, record.ID, executorID, "succeeded", JobSucceeded, artifactID, now)
	})
}

func (s *MySQLStore) mutateJob(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, now time.Time, mutate func(*sql.Tx, *JobRecord) error) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := scanJob(tx.QueryRowContext(ctx, jobSelect+` WHERE id = ? FOR UPDATE`, jobID).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if err := checkLease(record, executorID, leaseHash, generation, now); err != nil {
		return JobRecord{}, err
	}
	if err := mutate(tx, &record); err != nil {
		return JobRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func appendJobEvent(ctx context.Context, tx *sql.Tx, jobID, executorID, eventType string, state JobState, detail string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO local_executor_job_events(job_id, executor_id, event_type, state, detail, created_at)
VALUES (?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?)`, jobID, executorID, eventType, state, bounded(detail, 512), now)
	return err
}

func scanJob(scan scanFunc) (JobRecord, error) {
	var record JobRecord
	var payload []byte
	var state string
	var leaseExecutor sql.NullString
	var leaseHash []byte
	var leaseExpires sql.NullTime
	var acceptedAt sql.NullTime
	var acceptedAccount sql.NullString
	var submissionID sql.NullString
	var artifactID sql.NullString
	var errorCode sql.NullString
	var errorMessage sql.NullString
	if err := scan(
		&record.ID, &record.OwnerUsername, &record.SourceTaskID, &record.Platform, &payload, &state,
		&record.CancelRequested, &leaseExecutor, &leaseHash, &record.LeaseGeneration, &leaseExpires,
		&acceptedAt, &acceptedAccount, &submissionID, &artifactID, &errorCode, &errorMessage,
		&record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		return JobRecord{}, err
	}
	record.Payload = append(record.Payload[:0], payload...)
	record.State = JobState(state)
	if leaseExecutor.Valid {
		record.LeaseExecutorID = leaseExecutor.String
	}
	if len(leaseHash) == len(record.LeaseTokenHash) {
		copy(record.LeaseTokenHash[:], leaseHash)
	}
	if leaseExpires.Valid {
		value := leaseExpires.Time
		record.LeaseExpiresAt = &value
	}
	if acceptedAt.Valid {
		value := acceptedAt.Time
		record.AcceptedAt = &value
	}
	if acceptedAccount.Valid {
		record.AcceptedAccountID = acceptedAccount.String
	}
	if submissionID.Valid {
		record.SubmissionID = submissionID.String
	}
	if artifactID.Valid {
		record.ArtifactID = artifactID.String
	}
	if errorCode.Valid {
		record.ErrorCode = errorCode.String
	}
	if errorMessage.Valid {
		record.ErrorMessage = errorMessage.String
	}
	return record, nil
}
