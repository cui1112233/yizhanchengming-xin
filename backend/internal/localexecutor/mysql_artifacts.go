package localexecutor

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const artifactSelect = `SELECT id, job_id, owner_username, media_type, byte_size, sha256, storage_ref, created_at FROM local_executor_artifacts`

func (s *MySQLStore) CreateArtifact(ctx context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, input ArtifactRecord, now time.Time) (ArtifactRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ArtifactRecord{}, err
	}
	defer tx.Rollback()

	job, err := scanJob(tx.QueryRowContext(ctx, jobSelect+` WHERE id = ? FOR UPDATE`, jobID).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return ArtifactRecord{}, ErrJobNotFound
	}
	if err != nil {
		return ArtifactRecord{}, err
	}
	if err := checkLease(job, executorID, leaseHash, generation, now); err != nil {
		return ArtifactRecord{}, err
	}
	if !isAccepted(job) || job.State != JobUploading {
		return ArtifactRecord{}, ErrInvalidJobState
	}

	if _, err := scanArtifact(tx.QueryRowContext(ctx, artifactSelect+` WHERE id = ? FOR UPDATE`, input.ID).Scan); err == nil {
		return ArtifactRecord{}, ErrArtifactConflict
	} else if !errors.Is(err, sql.ErrNoRows) {
		return ArtifactRecord{}, err
	}

	input.JobID = jobID
	input.OwnerUsername = job.OwnerUsername
	input.CreatedAt = now
	_, err = tx.ExecContext(ctx, `INSERT INTO local_executor_artifacts
(id, job_id, owner_username, media_type, byte_size, sha256, storage_ref, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ID, input.JobID, input.OwnerUsername, input.MediaType, input.ByteSize, input.SHA256, input.StorageRef, input.CreatedAt)
	if err != nil {
		return ArtifactRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return ArtifactRecord{}, err
	}
	return input, nil
}

func (s *MySQLStore) ArtifactForOwner(ctx context.Context, owner, id string) (ArtifactRecord, error) {
	record, err := scanArtifact(s.db.QueryRowContext(ctx, artifactSelect+` WHERE id = ? AND owner_username = ?`, id, owner).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return ArtifactRecord{}, ErrArtifactNotFound
	}
	return record, err
}

func (s *MySQLStore) ArtifactForJob(ctx context.Context, jobID string) (ArtifactRecord, error) {
	record, err := scanArtifact(s.db.QueryRowContext(ctx, artifactSelect+` WHERE job_id = ? ORDER BY created_at, id LIMIT 1`, jobID).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return ArtifactRecord{}, ErrArtifactNotFound
	}
	return record, err
}

func (s *MySQLStore) ArtifactsForJob(ctx context.Context, jobID string) ([]ArtifactRecord, error) {
	rows, err := s.db.QueryContext(ctx, artifactSelect+` WHERE job_id = ? ORDER BY created_at, id`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []ArtifactRecord
	for rows.Next() {
		record, err := scanArtifact(rows.Scan)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, ErrArtifactNotFound
	}
	return records, nil
}

func scanArtifact(scan scanFunc) (ArtifactRecord, error) {
	var record ArtifactRecord
	var byteSize sql.NullInt64
	var sha256Value sql.NullString
	var storageRef sql.NullString
	if err := scan(
		&record.ID, &record.JobID, &record.OwnerUsername, &record.MediaType,
		&byteSize, &sha256Value, &storageRef, &record.CreatedAt,
	); err != nil {
		return ArtifactRecord{}, err
	}
	if byteSize.Valid {
		record.ByteSize = byteSize.Int64
	}
	if sha256Value.Valid {
		record.SHA256 = sha256Value.String
	}
	if storageRef.Valid {
		record.StorageRef = storageRef.String
	}
	return record, nil
}
