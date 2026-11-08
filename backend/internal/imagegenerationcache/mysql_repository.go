package imagegenerationcache

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }

func (r *MySQLRepository) Claim(ctx context.Context, owner, fingerprint, modelID string, lease time.Duration) (Entry, string, error) {
	owner, fingerprint, modelID = strings.TrimSpace(owner), strings.TrimSpace(fingerprint), strings.TrimSpace(modelID)
	if r == nil || r.db == nil || owner == "" || fingerprint == "" || modelID == "" {
		return Entry{}, "", fmt.Errorf("invalid image generation cache claim")
	}
	if lease <= 0 {
		lease = 10 * time.Minute
	}
	leaseUntil := time.Now().Add(lease)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, "", err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO batch_factory_v12_image_generation_cache(owner_username,fingerprint,model_id,state,lease_until,created_at,updated_at) VALUES(?,?,?,'generating',?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6))`, owner, fingerprint, modelID, leaseUntil)
	if err != nil {
		return Entry{}, "", err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		id, _ := result.LastInsertId()
		if err := tx.Commit(); err != nil {
			return Entry{}, "", err
		}
		return Entry{ID: id, State: StateGenerating, LeaseUntil: leaseUntil}, ClaimAcquired, nil
	}

	var entry Entry
	var leaseValue sql.NullTime
	var queueMS, headerMS, bodyMS, downloadMS, saveMS, totalMS int64
	err = tx.QueryRowContext(ctx, `SELECT id,state,COALESCE(storage_ref,''),COALESCE(media_type,''),lease_until,queue_ms,header_ms,body_ms,download_ms,save_ms,total_ms FROM batch_factory_v12_image_generation_cache WHERE owner_username=? AND fingerprint=? FOR UPDATE`, owner, fingerprint).
		Scan(&entry.ID, &entry.State, &entry.StorageRef, &entry.MediaType, &leaseValue, &queueMS, &headerMS, &bodyMS, &downloadMS, &saveMS, &totalMS)
	if err != nil {
		return Entry{}, "", err
	}
	entry.Timing = Timing{Queue: time.Duration(queueMS) * time.Millisecond, Header: time.Duration(headerMS) * time.Millisecond, Body: time.Duration(bodyMS) * time.Millisecond, Download: time.Duration(downloadMS) * time.Millisecond, Save: time.Duration(saveMS) * time.Millisecond, Total: time.Duration(totalMS) * time.Millisecond}
	if leaseValue.Valid {
		entry.LeaseUntil = leaseValue.Time
	}
	if entry.State == StateReady && entry.StorageRef != "" {
		if err := tx.Commit(); err != nil {
			return Entry{}, "", err
		}
		return entry, ClaimReady, nil
	}
	if entry.State == StateGenerating && entry.LeaseUntil.After(time.Now()) {
		if err := tx.Commit(); err != nil {
			return Entry{}, "", err
		}
		return entry, ClaimBusy, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_image_generation_cache SET model_id=?,state='generating',storage_ref=NULL,media_type=NULL,lease_until=?,last_error=NULL,updated_at=CURRENT_TIMESTAMP(6) WHERE id=?`, modelID, leaseUntil, entry.ID); err != nil {
		return Entry{}, "", err
	}
	if err := tx.Commit(); err != nil {
		return Entry{}, "", err
	}
	entry.State, entry.LeaseUntil, entry.StorageRef, entry.MediaType = StateGenerating, leaseUntil, "", ""
	return entry, ClaimAcquired, nil
}

func (r *MySQLRepository) MarkReady(ctx context.Context, owner, fingerprint, storageRef, mediaType string, timing Timing) error {
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_image_generation_cache SET state='ready',storage_ref=?,media_type=?,lease_until=NULL,queue_ms=?,header_ms=?,body_ms=?,download_ms=?,save_ms=?,total_ms=?,last_error=NULL,updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND fingerprint=? AND state='generating'`, storageRef, mediaType, timing.Queue.Milliseconds(), timing.Header.Milliseconds(), timing.Body.Milliseconds(), timing.Download.Milliseconds(), timing.Save.Milliseconds(), timing.Total.Milliseconds(), owner, fingerprint)
	return err
}

func (r *MySQLRepository) MarkFailed(ctx context.Context, owner, fingerprint string, cause error) error {
	message := "image generation failed"
	if cause != nil {
		message = cause.Error()
	}
	if len(message) > 512 {
		message = message[:512]
	}
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_image_generation_cache SET state='failed',storage_ref=NULL,media_type=NULL,lease_until=NULL,last_error=?,updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND fingerprint=? AND state IN ('generating','ready')`, message, owner, fingerprint)
	return err
}
