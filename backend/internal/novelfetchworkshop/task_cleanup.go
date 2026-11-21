package novelfetchworkshop

import (
	"context"
	"strings"
	"time"
)

// 批量测书任务本身没有“作品”实体，但设置页的「作品保留时长」同样适用于
// 已结束的测书任务：到期后删除任务文档与运行记录，正文转入可回收状态后
// 继续走 BodyRetentionDays 的延迟回收，避免误删后无法恢复。

// novelFetchTerminalStatusSQL 只包含已经结束、不会再被自动队列接管的
// 状态。等待重试、排队、处理中、待确认上传等状态即使超时也不在这里删除，
// 否则会把仍需人工处理的问题任务清掉。
const novelFetchTerminalStatusSQL = `LOWER(COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.status')),''), JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.queueState')))) IN ('completed','failed','stopped','cancelled','canceled','ai_failed','original_failed')`

type TaskCleanupRecord struct {
	Owner  string `json:"owner"`
	BookID string `json:"bookId"`
}

type TaskCleanupResult struct {
	Deleted     int                 `json:"deleted"`
	RunsDeleted int                 `json:"runsDeleted"`
	Results     []TaskCleanupRecord `json:"results"`
}

func normalizeTaskCleanupLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

// novelFetchTaskRetentionExpiry 与作品保留保持一致：最后活动日 + N 天的
// 本地零点。例如 11-14 任意时刻结束、保留 7 天，11-21 00:00 后到期。
func novelFetchTaskRetentionExpiry(lastActivity time.Time, days int, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	local := lastActivity.In(loc)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return start.AddDate(0, 0, days).UTC()
}

// ListDocumentOwners 返回当前仍有测书任务的账号，供按账号保留时长巡检。
func (s *MySQLStore) ListDocumentOwners(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT owner_username FROM novel_fetch_workshop_documents ORDER BY owner_username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	owners := []string{}
	for rows.Next() {
		var owner string
		if err := rows.Scan(&owner); err != nil {
			return nil, err
		}
		if strings.TrimSpace(owner) != "" {
			owners = append(owners, owner)
		}
	}
	return owners, rows.Err()
}

type mysqlTaskCleanupCandidate struct {
	bookID  string
	updated time.Time
}

// CleanupExpiredTasks 删除某账号下已到保留期且处于终态的测书任务，并清掉
// 同一保留期之前的运行记录。删除任务复用 DeleteDocuments 的语义：正文先
// 转为 releasable，再由 CleanupExpiredBodies 在 BodyRetentionDays 后真正删除。
func (s *MySQLStore) CleanupExpiredTasks(ctx context.Context, owner string, now time.Time, retentionDays, limit int) (TaskCleanupResult, error) {
	result := TaskCleanupResult{Results: []TaskCleanupRecord{}}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return result, nil
	}
	now = normalizeCleanupNow(now)
	days := NormalizeBodyRetentionDays(retentionDays)
	limit = normalizeTaskCleanupLimit(limit)

	// 运行记录没有稳定的书号关联（批次级 run 的 book_id 存的是批次号），
	// 因此按账号 + 时间 + 非活动状态清理，避免误删仍在排队/执行的记录。
	runCutoff := now.Add(-time.Duration(days) * 24 * time.Hour)
	runResult, err := s.DB.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_runs WHERE owner_username=? AND updated_at<=? AND status NOT IN ('queued','running')`, owner, runCutoff)
	if err != nil {
		return result, err
	}
	if affected, rowsErr := runResult.RowsAffected(); rowsErr == nil {
		result.RunsDeleted = int(affected)
	}

	// 预筛只取“按自然日可能已到期”的任务；精确到期时间在 Go 侧按本地零点
	// 计算，避免数据库时区与账号展示时区不一致。
	prefilter := now.Add(-time.Duration(days-1) * 24 * time.Hour)
	rows, err := s.DB.QueryContext(ctx, `SELECT book_id, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND updated_at<=? AND `+novelFetchTerminalStatusSQL+` ORDER BY updated_at ASC LIMIT ?`, owner, prefilter, limit)
	if err != nil {
		return result, err
	}
	candidates := make([]mysqlTaskCleanupCandidate, 0, limit)
	for rows.Next() {
		var candidate mysqlTaskCleanupCandidate
		if err := rows.Scan(&candidate.bookID, &candidate.updated); err != nil {
			rows.Close()
			return result, err
		}
		candidate.updated = candidate.updated.UTC()
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()

	for _, candidate := range candidates {
		if now.Before(novelFetchTaskRetentionExpiry(candidate.updated, days, time.Local)) {
			// 已按更新时间升序排列，后面的任务只会更新，不会更早到期。
			break
		}
		if err := s.deleteExpiredTask(ctx, owner, candidate.bookID, now); err != nil {
			return result, err
		}
		result.Deleted++
		result.Results = append(result.Results, TaskCleanupRecord{Owner: owner, BookID: candidate.bookID})
	}
	return result, nil
}

func (s *MySQLStore) deleteExpiredTask(ctx context.Context, owner, bookID string, releasedAt time.Time) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	expiresAt := releasedAt.Add(DefaultBodyRetentionDays * 24 * time.Hour)
	if _, err := tx.ExecContext(ctx, `UPDATE novel_fetch_workshop_bodies SET state='releasable', releasable_at=?, expires_at=?, updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=? AND state<>'releasable'`, releasedAt, expiresAt, owner, bookID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`, owner, bookID); err != nil {
		return err
	}
	return tx.Commit()
}
