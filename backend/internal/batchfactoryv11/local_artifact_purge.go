package batchfactoryv11

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"qiantie/backend/internal/localartifact"
)

var localMergedArtifactID = regexp.MustCompile(`^merge_[A-Za-z0-9_-]{3,96}$`)

const DefaultLocalMergedArtifactRetentionDays = 3

func normalizeLocalMergedArtifactRetentionDays(value int) int {
	switch value {
	case 3, 7, 14, 30:
		return value
	default:
		return DefaultLocalMergedArtifactRetentionDays
	}
}

// localMergedArtifactStorageRef returns the root-confined file name for a
// completed local merge URL. Remote URLs, another batch's URL and anything
// that is not a generated merge artifact are deliberately not eligible for
// local retention cleanup.
func localMergedArtifactStorageRef(batchID, outputURL string) (string, bool) {
	batchID = strings.TrimSpace(batchID)
	if batchID == "" {
		return "", false
	}
	prefix := "/api/batch-factory/v12/batches/" + url.PathEscape(batchID) + "/merge-media/"
	artifactID, ok := strings.CutPrefix(strings.TrimSpace(outputURL), prefix)
	if !ok || !localMergedArtifactID.MatchString(artifactID) {
		return "", false
	}
	return artifactID + ".mp4", true
}

func isLocalMergedArtifactStorageRef(storageRef string) bool {
	storageRef = strings.TrimSpace(storageRef)
	if !strings.HasSuffix(storageRef, ".mp4") {
		return false
	}
	return localMergedArtifactID.MatchString(strings.TrimSuffix(storageRef, ".mp4"))
}

type LocalArtifactPurge struct {
	ID         int64
	StorageRef string
}

type LocalArtifactPurgeRepository interface {
	ClaimDueLocalArtifactPurge(context.Context) (LocalArtifactPurge, bool, error)
	MarkLocalArtifactPurgeDeleted(context.Context, int64) error
	MarkLocalArtifactPurgeRetry(context.Context, int64, error) error
}

type LocalArtifactPurgeSummary struct {
	Deleted int
	Retried int
}

type LocalArtifactPurger struct {
	Repository LocalArtifactPurgeRepository
	Artifacts  *localartifact.Store
}

// RunOnce drains the currently due purge ledger. A failed individual file is
// recorded as retryable; a repository failure stops the run so it cannot hide
// a database outage behind a successful cleanup report.
func (p LocalArtifactPurger) RunOnce(ctx context.Context) (LocalArtifactPurgeSummary, error) {
	summary := LocalArtifactPurgeSummary{}
	if p.Repository == nil || p.Artifacts == nil {
		return summary, fmt.Errorf("local artifact purger is not configured")
	}
	for {
		item, found, err := p.Repository.ClaimDueLocalArtifactPurge(ctx)
		if err != nil {
			return summary, err
		}
		if !found {
			return summary, nil
		}
		if !isLocalMergedArtifactStorageRef(item.StorageRef) {
			err = fmt.Errorf("invalid local merged artifact reference %q", item.StorageRef)
		} else {
			err = p.Artifacts.Remove(item.StorageRef)
		}
		if err == nil {
			if markErr := p.Repository.MarkLocalArtifactPurgeDeleted(ctx, item.ID); markErr != nil {
				return summary, markErr
			}
			summary.Deleted++
			continue
		}
		if markErr := p.Repository.MarkLocalArtifactPurgeRetry(ctx, item.ID, err); markErr != nil {
			return summary, markErr
		}
		summary.Retried++
	}
}

// Start launches a bounded periodic scan. The initial pass makes recovery
// deterministic after a container restart; claimed rows may be reclaimed
// after fifteen minutes if a process dies between claim and completion.
func (p LocalArtifactPurger) Start(interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	go func() {
		_, _ = p.RunOnce(context.Background())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			_, _ = p.RunOnce(context.Background())
		}
	}()
}

func (s *MySQLStore) ClaimDueLocalArtifactPurge(ctx context.Context) (LocalArtifactPurge, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LocalArtifactPurge{}, false, err
	}
	defer tx.Rollback()
	var item LocalArtifactPurge
	err = tx.QueryRowContext(ctx, `SELECT id,storage_ref FROM batch_factory_v11_local_artifact_purges WHERE purge_after<=CURRENT_TIMESTAMP(6) AND (state IN ('pending','retry') OR (state='processing' AND updated_at<=DATE_SUB(CURRENT_TIMESTAMP(6), INTERVAL 15 MINUTE))) ORDER BY purge_after,id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&item.ID, &item.StorageRef)
	if err == sql.ErrNoRows {
		if commitErr := tx.Commit(); commitErr != nil {
			return LocalArtifactPurge{}, false, commitErr
		}
		return LocalArtifactPurge{}, false, nil
	}
	if err != nil {
		return LocalArtifactPurge{}, false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v11_local_artifact_purges SET state='processing',attempt_count=attempt_count+1,updated_at=CURRENT_TIMESTAMP(6) WHERE id=?`, item.ID); err != nil {
		return LocalArtifactPurge{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return LocalArtifactPurge{}, false, err
	}
	return item, true, nil
}

func (s *MySQLStore) MarkLocalArtifactPurgeDeleted(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v11_local_artifact_purges SET state='deleted',last_error=NULL,deleted_at=CURRENT_TIMESTAMP(6),updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='processing'`, id)
	return err
}

func (s *MySQLStore) MarkLocalArtifactPurgeRetry(ctx context.Context, id int64, cause error) error {
	message := "local artifact cleanup failed"
	if cause != nil {
		message = cause.Error()
	}
	if len(message) > 512 {
		message = message[:512]
	}
	// Move the retry window forward. Without this, a permanently invalid file
	// reference would be claimed again immediately and turn one periodic scan
	// into a tight database loop.
	_, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v11_local_artifact_purges SET state='retry',last_error=?,purge_after=DATE_ADD(CURRENT_TIMESTAMP(6), INTERVAL 15 MINUTE),updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND state='processing'`, message, id)
	return err
}

// queueLocalMergedArtifactPurges snapshots only local completed merge output
// before its merge-job rows are removed. It deliberately ignores remote URLs:
// this local retention worker has no authority to remove TOS objects.
func queueLocalMergedArtifactPurges(ctx context.Context, tx *sql.Tx, owner, batchID, bookID string, purgeAfter time.Time) error {
	query := `SELECT COALESCE(book_id,''),COALESCE(output_url,'') FROM batch_factory_v11_merge_jobs WHERE owner_username=? AND batch_id=?`
	args := []any{owner, batchID}
	if strings.TrimSpace(bookID) != "" {
		query += ` AND book_id=?`
		args = append(args, bookID)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var mergedBookID, outputURL string
		if err := rows.Scan(&mergedBookID, &outputURL); err != nil {
			return err
		}
		storageRef, ok := localMergedArtifactStorageRef(batchID, outputURL)
		if !ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO batch_factory_v11_local_artifact_purges(owner_username,batch_id,book_id,storage_ref,purge_after,state,attempt_count,created_at,updated_at) VALUES(?,?,?,?,?,'pending',0,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6))`, owner, batchID, mergedBookID, storageRef, purgeAfter); err != nil {
			return err
		}
	}
	return rows.Err()
}
