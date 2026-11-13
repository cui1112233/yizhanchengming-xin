package productionmedia

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }

func (r *MySQLRepository) Enqueue(ctx context.Context, item Object) error {
	item.Owner = strings.TrimSpace(item.Owner)
	item.LocalStorageRef = strings.TrimSpace(item.LocalStorageRef)
	item.ObjectKey = strings.Trim(strings.TrimSpace(item.ObjectKey), "/")
	item.MediaType = strings.TrimSpace(item.MediaType)
	if r == nil || r.db == nil || item.Owner == "" || filepath.Base(item.LocalStorageRef) != item.LocalStorageRef || item.LocalStorageRef == "" || item.ObjectKey == "" || strings.Contains(item.ObjectKey, "..") {
		return fmt.Errorf("invalid production media enqueue")
	}
	if item.MediaType == "" {
		item.MediaType = "application/octet-stream"
	}
	// Scanner 每分钟都会把仍在本地缓存里的对象重新入队一次。这里只有在
	// media_type / retention_days 真正变化时才写行；否则 ON DUPLICATE KEY
	// UPDATE 产生 0 行变更，不会向 binlog 追加无意义记录。
	_, err := r.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_media_objects(owner_username,media_kind,media_type,local_storage_ref,object_key,state,attempt_count,retention_days,created_at,updated_at) VALUES(?,?,?,?,?,'pending',0,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE media_type=IF(media_type<>VALUES(media_type),VALUES(media_type),media_type),retention_days=IF(retention_days<>VALUES(retention_days),VALUES(retention_days),retention_days)`, item.Owner, item.MediaKind, item.MediaType, item.LocalStorageRef, item.ObjectKey, NormalizeRetentionDays(item.RetentionDays))
	return err
}

func (r *MySQLRepository) ClaimUpload(ctx context.Context) (Object, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Object{}, false, err
	}
	defer tx.Rollback()
	var item Object
	err = tx.QueryRowContext(ctx, `SELECT id,owner_username,media_kind,media_type,local_storage_ref,object_key,retention_days FROM batch_factory_v12_media_objects WHERE state='pending' OR (state='failed' AND updated_at<=DATE_SUB(CURRENT_TIMESTAMP(6),INTERVAL 5 MINUTE)) OR (state='uploading' AND updated_at<=DATE_SUB(CURRENT_TIMESTAMP(6),INTERVAL 15 MINUTE)) ORDER BY updated_at,id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&item.ID, &item.Owner, &item.MediaKind, &item.MediaType, &item.LocalStorageRef, &item.ObjectKey, &item.RetentionDays)
	if err == sql.ErrNoRows {
		if err := tx.Commit(); err != nil {
			return Object{}, false, err
		}
		return Object{}, false, nil
	}
	if err != nil {
		return Object{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='uploading',attempt_count=attempt_count+1,last_error=NULL,updated_at=CURRENT_TIMESTAMP(6) WHERE id=?`, item.ID); err != nil {
		return Object{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Object{}, false, err
	}
	item.State = StateUploading
	return item, true, nil
}

func (r *MySQLRepository) MarkVerified(ctx context.Context, item Object) error {
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET byte_size=?,sha256=?,state='verified',last_error=NULL,remote_verified_at=?,purge_after=?,local_deleted_at=NULL,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='uploading'`, item.ByteSize, item.SHA256, item.RemoteVerifiedAt, item.PurgeAfter, item.ID)
	return err
}

func (r *MySQLRepository) MarkUploadFailed(ctx context.Context, id int64, cause error) error {
	message := boundedError(cause)
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='failed',last_error=?,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='uploading'`, message, id)
	return err
}

func (r *MySQLRepository) FindVerifiedByStorageRef(ctx context.Context, storageRef string) (Object, bool, error) {
	var item Object
	err := r.db.QueryRowContext(ctx, `SELECT id,owner_username,media_kind,media_type,local_storage_ref,object_key,byte_size,sha256,state,retention_days,remote_verified_at,purge_after FROM batch_factory_v12_media_objects WHERE local_storage_ref=? AND remote_verified_at IS NOT NULL AND state IN ('verified','cleanup_pending','cleaned') ORDER BY id DESC LIMIT 1`, storageRef).Scan(&item.ID, &item.Owner, &item.MediaKind, &item.MediaType, &item.LocalStorageRef, &item.ObjectKey, &item.ByteSize, &item.SHA256, &item.State, &item.RetentionDays, &item.RemoteVerifiedAt, &item.PurgeAfter)
	if err == sql.ErrNoRows {
		return Object{}, false, nil
	}
	return item, err == nil, err
}

func (r *MySQLRepository) MarkRehydrated(ctx context.Context, id int64, purgeAfter time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='verified',local_deleted_at=NULL,purge_after=?,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND remote_verified_at IS NOT NULL`, purgeAfter, id)
	return err
}

func (r *MySQLRepository) ClaimDuePurge(ctx context.Context, pressure bool) (Object, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Object{}, false, err
	}
	defer tx.Rollback()
	condition := `purge_after<=CURRENT_TIMESTAMP(6)`
	if pressure {
		condition = `1=1`
	}
	query := `SELECT id,owner_username,media_kind,media_type,local_storage_ref,object_key,byte_size,sha256,state,retention_days,remote_verified_at,purge_after FROM batch_factory_v12_media_objects WHERE remote_verified_at IS NOT NULL AND local_deleted_at IS NULL AND state IN ('verified','cleanup_pending') AND (` + condition + `) ORDER BY remote_verified_at,id LIMIT 1 FOR UPDATE SKIP LOCKED`
	var item Object
	err = tx.QueryRowContext(ctx, query).Scan(&item.ID, &item.Owner, &item.MediaKind, &item.MediaType, &item.LocalStorageRef, &item.ObjectKey, &item.ByteSize, &item.SHA256, &item.State, &item.RetentionDays, &item.RemoteVerifiedAt, &item.PurgeAfter)
	if err == sql.ErrNoRows {
		if err := tx.Commit(); err != nil {
			return Object{}, false, err
		}
		return Object{}, false, nil
	}
	if err != nil {
		return Object{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='cleanup_pending',updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND remote_verified_at IS NOT NULL`, item.ID); err != nil {
		return Object{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Object{}, false, err
	}
	item.State = StateCleanupPending
	return item, true, nil
}

func (r *MySQLRepository) MarkLocalDeleted(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='cleaned',local_deleted_at=CURRENT_TIMESTAMP(6),last_error=NULL,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='cleanup_pending' AND remote_verified_at IS NOT NULL`, id)
	return err
}

func (r *MySQLRepository) MarkPurgeRetry(ctx context.Context, id int64, cause error) error {
	_, err := r.db.ExecContext(ctx, `UPDATE batch_factory_v12_media_objects SET state='verified',last_error=?,purge_after=DATE_ADD(CURRENT_TIMESTAMP(6),INTERVAL 15 MINUTE),updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='cleanup_pending' AND remote_verified_at IS NOT NULL`, boundedError(cause), id)
	return err
}

func boundedError(cause error) string {
	message := "production media operation failed"
	if cause != nil {
		message = cause.Error()
	}
	if len(message) > 512 {
		message = message[:512]
	}
	return message
}
