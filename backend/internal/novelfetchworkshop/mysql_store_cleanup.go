package novelfetchworkshop

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

func (s *MySQLStore) MarkBodyReleasable(ctx context.Context, owner, bookID, versionID string, releasableAt time.Time, retentionDays int) (BodyRef, error) {
	bookID = strings.TrimSpace(bookID)
	versionID = strings.TrimSpace(versionID)
	if bookID == "" || versionID == "" {
		return BodyRef{}, ErrNotFound
	}
	released := normalizeCleanupNow(releasableAt)
	expires := released.Add(time.Duration(NormalizeBodyRetentionDays(retentionDays)) * 24 * time.Hour)
	result, err := s.DB.ExecContext(ctx, `UPDATE novel_fetch_workshop_bodies SET state='releasable', releasable_at=?, expires_at=? WHERE owner_username=? AND book_id=? AND version_id=?`, released, expires, owner, bookID, versionID)
	if err != nil {
		return BodyRef{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return BodyRef{}, err
	}
	if affected == 0 {
		return BodyRef{}, ErrNotFound
	}

	ref := BodyRef{VersionID: versionID}
	var updated time.Time
	var lastNeeded time.Time
	var storedReleased time.Time
	var storedExpires time.Time
	err = s.DB.QueryRowContext(ctx, `SELECT revision, content_hash, char_count, state, updated_at, last_needed_at, releasable_at, expires_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=?`, owner, bookID, versionID).
		Scan(&ref.Revision, &ref.ContentHash, &ref.CharCount, &ref.State, &updated, &lastNeeded, &storedReleased, &storedExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return BodyRef{}, ErrNotFound
	}
	if err != nil {
		return BodyRef{}, err
	}
	ref.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	ref.LastNeededAt = lastNeeded.UTC().Format(time.RFC3339Nano)
	ref.ReleasableAt = storedReleased.UTC().Format(time.RFC3339Nano)
	ref.ExpiresAt = storedExpires.UTC().Format(time.RFC3339Nano)
	return ref, nil
}

type mysqlCleanupCandidate struct {
	bookID       string
	versionID    string
	encoding     string
	blob         []byte
	releasableAt time.Time
}

func (s *MySQLStore) CleanupBodies(ctx context.Context, owner string, request BodyCleanupRequest) (BodyCleanupResult, error) {
	if request.Reason != BodyCleanupReasonExpired && request.Reason != BodyCleanupReasonCapacity {
		return BodyCleanupResult{}, errors.New("invalid body cleanup reason")
	}
	limit := normalizeCleanupLimit(request.Limit)
	var (
		rows *sql.Rows
		err  error
	)
	if request.Reason == BodyCleanupReasonExpired {
		now := normalizeCleanupNow(request.Now)
		rows, err = s.DB.QueryContext(ctx, `SELECT book_id, version_id, content_encoding, content_blob, releasable_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND state='releasable' AND releasable_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at<=? ORDER BY releasable_at ASC, book_id ASC, version_id ASC LIMIT ?`, owner, now, limit)
	} else {
		rows, err = s.DB.QueryContext(ctx, `SELECT book_id, version_id, content_encoding, content_blob, releasable_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND state='releasable' AND releasable_at IS NOT NULL ORDER BY releasable_at ASC, book_id ASC, version_id ASC LIMIT ?`, owner, limit)
	}
	if err != nil {
		return BodyCleanupResult{}, err
	}

	candidates := make([]mysqlCleanupCandidate, 0)
	for rows.Next() {
		var candidate mysqlCleanupCandidate
		if err := rows.Scan(&candidate.bookID, &candidate.versionID, &candidate.encoding, &candidate.blob, &candidate.releasableAt); err != nil {
			_ = rows.Close()
			return BodyCleanupResult{}, err
		}
		candidate.releasableAt = candidate.releasableAt.UTC()
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return BodyCleanupResult{}, err
	}
	if err := rows.Close(); err != nil {
		return BodyCleanupResult{}, err
	}
	result := BodyCleanupResult{Results: make([]BodyCleanupRecord, 0, len(candidates))}
	if len(candidates) == 0 {
		return result, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return BodyCleanupResult{}, err
	}
	defer tx.Rollback()
	for _, candidate := range candidates {
		if s.ObjectBodies != nil && strings.EqualFold(strings.TrimSpace(candidate.encoding), "tos") {
			if objectKey := strings.TrimSpace(string(candidate.blob)); objectKey != "" {
				if err := s.ObjectBodies.DeleteContent(ctx, objectKey); err != nil {
					return BodyCleanupResult{}, err
				}
			}
		}
		deleted, err := tx.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=? AND state='releasable' AND releasable_at=?`, owner, candidate.bookID, candidate.versionID, candidate.releasableAt)
		if err != nil {
			return BodyCleanupResult{}, err
		}
		affected, err := deleted.RowsAffected()
		if err != nil {
			return BodyCleanupResult{}, err
		}
		if affected == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE novel_fetch_workshop_documents SET document_json=JSON_REMOVE(document_json, ?), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`, bodyRefJSONPath(candidate.versionID), owner, candidate.bookID); err != nil {
			return BodyCleanupResult{}, err
		}
		result.Deleted++
		result.Results = append(result.Results, BodyCleanupRecord{BookID: candidate.bookID, VersionID: candidate.versionID, Reason: request.Reason})
	}
	if err := tx.Commit(); err != nil {
		return BodyCleanupResult{}, err
	}
	return result, nil
}

// CleanupExpiredBodies 跨账号回收已到期的可释放正文，供后台定时任务调用。
// 普通 CleanupBodies 只处理单个账号，删除历史文档后如果没有人手动点清理，
// 正文会永久占用 TOS 容量。
func (s *MySQLStore) CleanupExpiredBodies(ctx context.Context, now time.Time, limit int) (BodyCleanupResult, error) {
	now = normalizeCleanupNow(now)
	limit = normalizeCleanupLimit(limit)
	rows, err := s.DB.QueryContext(ctx, `SELECT owner_username, book_id, version_id, content_encoding, content_blob, releasable_at FROM novel_fetch_workshop_bodies WHERE state='releasable' AND releasable_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at<=? ORDER BY expires_at ASC, owner_username ASC, book_id ASC, version_id ASC LIMIT ?`, now, limit)
	if err != nil {
		return BodyCleanupResult{}, err
	}
	type cleanupCandidate struct {
		owner        string
		bookID       string
		versionID    string
		encoding     string
		blob         []byte
		releasableAt time.Time
	}
	candidates := make([]cleanupCandidate, 0)
	for rows.Next() {
		var candidate cleanupCandidate
		if err := rows.Scan(&candidate.owner, &candidate.bookID, &candidate.versionID, &candidate.encoding, &candidate.blob, &candidate.releasableAt); err != nil {
			_ = rows.Close()
			return BodyCleanupResult{}, err
		}
		candidate.releasableAt = candidate.releasableAt.UTC()
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return BodyCleanupResult{}, err
	}
	if err := rows.Close(); err != nil {
		return BodyCleanupResult{}, err
	}
	result := BodyCleanupResult{Results: make([]BodyCleanupRecord, 0, len(candidates))}
	if len(candidates) == 0 {
		return result, nil
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return BodyCleanupResult{}, err
	}
	defer tx.Rollback()
	for _, candidate := range candidates {
		// 先删对象再删行；对象删除失败时保留数据库记录，下一轮仍可重试。
		if s.ObjectBodies != nil && strings.EqualFold(strings.TrimSpace(candidate.encoding), "tos") {
			if objectKey := strings.TrimSpace(string(candidate.blob)); objectKey != "" {
				if err := s.ObjectBodies.DeleteContent(ctx, objectKey); err != nil {
					return BodyCleanupResult{}, err
				}
			}
		}
		deleted, err := tx.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=? AND state='releasable' AND releasable_at=?`, candidate.owner, candidate.bookID, candidate.versionID, candidate.releasableAt)
		if err != nil {
			return BodyCleanupResult{}, err
		}
		affected, err := deleted.RowsAffected()
		if err != nil {
			return BodyCleanupResult{}, err
		}
		if affected == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE novel_fetch_workshop_documents SET document_json=JSON_REMOVE(document_json, ?), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`, bodyRefJSONPath(candidate.versionID), candidate.owner, candidate.bookID); err != nil {
			return BodyCleanupResult{}, err
		}
		result.Deleted++
		result.Results = append(result.Results, BodyCleanupRecord{BookID: candidate.bookID, VersionID: candidate.versionID, Reason: BodyCleanupReasonExpired})
	}
	if err := tx.Commit(); err != nil {
		return BodyCleanupResult{}, err
	}
	return result, nil
}

func bodyRefJSONPath(versionID string) string {
	return "$.bodyRefs." + strconv.Quote(versionID)
}
