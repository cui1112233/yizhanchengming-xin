package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/localartifact"
)

const (
	// DefaultBookRetentionGrace 是作品到期隐藏后到真正删除之间的缓冲时间。
	DefaultBookRetentionGrace = 24 * time.Hour
	bookRetentionClaimWindow  = 15 * time.Minute
	bookRetentionMaxPerRun    = 50
)

// RemoteObjectDeleter 只暴露删除单个对象的能力，避免清理器拿到整份 TOS 客户端。
type RemoteObjectDeleter interface {
	DeleteObject(context.Context, string) error
}

// RetentionBook 是一条待硬删的作品记录。
type RetentionBook struct {
	ID      string
	BatchID string
	Owner   string
}

// RetentionRecomputeSummary 描述一次重算的隐藏 / 恢复数量。
type RetentionRecomputeSummary struct {
	Hidden   int64
	Restored int64
}

// BookRetentionSupervisor 按作品维度执行保留策略：先重算到期时间，再把到期
// 作品隐藏，最后在缓冲期结束后删除云端对象、本地文件和数据库记录。
type BookRetentionSupervisor struct {
	Store         *MySQLStore
	Local         *localartifact.Store
	Remote        RemoteObjectDeleter
	RetentionDays func(string) int
	Grace         time.Duration
	MaxPerRun     int
}

// RunOnce 执行一次完整巡检。
func (s BookRetentionSupervisor) RunOnce(ctx context.Context) error {
	if s.Store == nil {
		return fmt.Errorf("book retention store is not configured")
	}
	owners, err := s.Store.ListBookRetentionOwners(ctx)
	if err != nil {
		return err
	}
	for _, owner := range owners {
		days := DefaultLocalMergedArtifactRetentionDays
		if s.RetentionDays != nil {
			days = normalizeLocalMergedArtifactRetentionDays(s.RetentionDays(owner))
		}
		if _, err := s.Store.RecomputeBookRetention(ctx, owner, days); err != nil {
			return err
		}
	}
	max := s.MaxPerRun
	if max <= 0 {
		max = bookRetentionMaxPerRun
	}
	for i := 0; i < max; i++ {
		book, found, err := s.Store.ClaimDueRetentionBook(ctx)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		if err := s.Store.PurgeRetentionBook(ctx, book, s.Local, s.Remote); err != nil {
			log.Printf("book retention purge failed book=%s batch=%s owner=%s: %v", book.ID, book.BatchID, book.Owner, err)
		}
	}
	return nil
}

// StartDaily 启动立即补扫，每小时复查，并在指定时间执行日常巡检。
// 真正删除仍受24小时缓冲期和每轮删除数量限制保护。
func (s BookRetentionSupervisor) StartDaily(hour, minute int) {
	go runBookRetentionLoop(context.Background(), s.RunOnce, func() time.Duration {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		delay := time.Until(next)
		// Retry interrupted cleanup and apply changed settings within an hour.
		if delay > time.Hour {
			delay = time.Hour
		}
		log.Printf("book retention next sweep at %s", now.Add(delay).Format("2006-01-02 15:04:05"))
		return delay
	})
}

func runBookRetentionLoop(ctx context.Context, run func(context.Context) error, delay func() time.Duration) {
	for ctx.Err() == nil {
		if err := run(ctx); err != nil {
			log.Printf("book retention sweep failed: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
		timer := time.NewTimer(delay())
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

// ListBookRetentionOwners 返回仍有作品的账号，避免对空账号做无意义扫描。
func (s *MySQLStore) ListBookRetentionOwners(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT owner_username FROM batch_factory_v12_books`)
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
		owners = append(owners, owner)
	}
	return owners, rows.Err()
}

type retentionBookActivity struct {
	id          string
	completedAt time.Time
	existing    sql.NullTime
}

// RecomputeBookRetention 按当前保留天数重算该账号所有作品的到期时间。
// 它同时处理两件事：把未到期但已隐藏的作品恢复显示；把已到期的作品隐藏并
// 安排 24 小时后的硬删。每次巡检都会重算，因此修改天数会追溯已有作品。
func (s *MySQLStore) RecomputeBookRetention(ctx context.Context, owner string, days int) (RetentionRecomputeSummary, error) {
	summary := RetentionRecomputeSummary{}
	if strings.TrimSpace(owner) == "" {
		return summary, ErrInvalid
	}
	days = normalizeLocalMergedArtifactRetentionDays(days)
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,
GREATEST(
  b.updated_at,
  COALESCE((SELECT MAX(s.updated_at) FROM batch_factory_v12_book_stage_runs s WHERE s.book_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(m.updated_at) FROM batch_factory_v12_merge_jobs m WHERE m.book_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(p.updated_at) FROM batch_factory_v12_comic_pages p WHERE p.book_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(j.updated_at) FROM batch_factory_v12_production_jobs j WHERE j.book_id=b.id), b.updated_at)
) AS completed_at,
b.completed_at
FROM batch_factory_v12_books b WHERE b.owner_username=?`, owner)
	if err != nil {
		return summary, err
	}
	activities := []retentionBookActivity{}
	for rows.Next() {
		var item retentionBookActivity
		if err := rows.Scan(&item.id, &item.completedAt, &item.existing); err != nil {
			rows.Close()
			return summary, err
		}
		// 完成时间只前进不后退：删除漫画页或轮换视频都不应让作品“变新”。
		if item.existing.Valid && item.existing.Time.After(item.completedAt) {
			item.completedAt = item.existing.Time
		}
		activities = append(activities, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return summary, err
	}
	rows.Close()

	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return summary, err
	}
	defer tx.Rollback()
	for _, item := range activities {
		expires := bookRetentionExpiry(item.completedAt, days)
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET completed_at=?,retention_expires_at=? WHERE id=? AND owner_username=?`, item.completedAt, expires, item.id, owner); err != nil {
			return summary, err
		}
	}
	if result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET hidden_at=NULL,purge_after=NULL WHERE owner_username=? AND hidden_at IS NOT NULL AND purge_after>? AND retention_expires_at>?`, owner, now, now); err != nil {
		return summary, err
	} else if affected, err := result.RowsAffected(); err == nil {
		summary.Restored = affected
	}
	if result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET hidden_at=?,purge_after=? WHERE owner_username=? AND hidden_at IS NULL AND retention_expires_at IS NOT NULL AND retention_expires_at<=?`, now, now.Add(DefaultBookRetentionGrace), owner, now); err != nil {
		return summary, err
	} else if affected, err := result.RowsAffected(); err == nil {
		summary.Hidden = affected
	}
	if err := tx.Commit(); err != nil {
		return summary, err
	}
	return summary, nil
}

func bookRetentionExpiry(completed time.Time, days int) time.Time {
	if days <= 0 {
		days = DefaultLocalMergedArtifactRetentionDays
	}
	return completed.AddDate(0, 0, days)
}

// ClaimDueRetentionBook 原子领取一个已过缓冲期的作品。purge_after 会被推后，
// 这样进程在删除中途退出时不会立即被下一次巡检重复抢占。
func (s *MySQLStore) ClaimDueRetentionBook(ctx context.Context) (RetentionBook, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RetentionBook{}, false, err
	}
	defer tx.Rollback()
	var book RetentionBook
	// 清理时间统一按 UTC 写入，因此这里也必须用 UTC_TIMESTAMP 比较；数据库
	// 会话时区是本地时间时，CURRENT_TIMESTAMP 会提前 8 小时触发删除。
	err = tx.QueryRowContext(ctx, `SELECT id,batch_id,owner_username FROM batch_factory_v12_books WHERE hidden_at IS NOT NULL AND purge_after<=UTC_TIMESTAMP(6) ORDER BY purge_after,id LIMIT 1 FOR UPDATE SKIP LOCKED`).
		Scan(&book.ID, &book.BatchID, &book.Owner)
	if errors.Is(err, sql.ErrNoRows) {
		if commitErr := tx.Commit(); commitErr != nil {
			return RetentionBook{}, false, commitErr
		}
		return RetentionBook{}, false, nil
	}
	if err != nil {
		return RetentionBook{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET purge_after=? WHERE id=?`, time.Now().UTC().Add(bookRetentionClaimWindow), book.ID); err != nil {
		return RetentionBook{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return RetentionBook{}, false, err
	}
	return book, true, nil
}

// PurgeRetentionBook 先删云端与本地文件，全部成功后再删数据库记录。任何一步
// 失败都会保留记录并等待下一轮重试，避免出现“记录没了、文件还在”的孤儿空间。
func (s *MySQLStore) PurgeRetentionBook(ctx context.Context, book RetentionBook, local *localartifact.Store, remote RemoteObjectDeleter) error {
	if strings.TrimSpace(book.Owner) == "" || strings.TrimSpace(book.BatchID) == "" || strings.TrimSpace(book.ID) == "" {
		return ErrInvalid
	}
	if local == nil {
		return fmt.Errorf("book retention local store is not configured")
	}
	refs, tosKeys, err := s.bookStorageRefs(ctx, book.Owner, book.ID)
	if err != nil {
		return err
	}
	if remote == nil && len(tosKeys) > 0 {
		return fmt.Errorf("TOS is not configured; retain book and cloud references for retry")
	}
	if len(refs) > 0 {
		shared, err := s.visibleBookStorageRefs(ctx, book.Owner, book.ID)
		if err != nil {
			return err
		}
		for ref := range refs {
			if _, ok := shared[ref]; ok {
				continue
			}
			if err := s.purgeStorageRef(ctx, book.Owner, ref, local, remote); err != nil {
				return err
			}
		}
	}
	if remote != nil {
		for key := range tosKeys {
			if err := remote.DeleteObject(ctx, key); err != nil {
				return fmt.Errorf("delete TOS object %s: %w", key, err)
			}
		}
	}
	return s.deleteRetentionBookRows(ctx, book)
}

func (s *MySQLStore) purgeStorageRef(ctx context.Context, owner, ref string, local *localartifact.Store, remote RemoteObjectDeleter) error {
	if filepath.Base(ref) != ref || ref == "" || ref == "." {
		return nil
	}
	var objectKey string
	err := s.db.QueryRowContext(ctx, `SELECT object_key FROM batch_factory_v12_media_objects WHERE owner_username=? AND local_storage_ref=?`, owner, ref).Scan(&objectKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if objectKey != "" && remote == nil {
		return fmt.Errorf("TOS is not configured; retain local file and object ledger for retry")
	}
	if err == nil && objectKey != "" && remote != nil {
		if err := remote.DeleteObject(ctx, objectKey); err != nil {
			return fmt.Errorf("delete TOS object %s: %w", objectKey, err)
		}
	}
	if err := local.Remove(ref); err != nil {
		return fmt.Errorf("delete local artifact %s: %w", ref, err)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM batch_factory_v12_media_objects WHERE owner_username=? AND local_storage_ref=?`, owner, ref); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM local_executor_artifacts WHERE owner_username=? AND storage_ref=?`, owner, ref); err != nil {
		return err
	}
	return nil
}

// bookStorageRefs 汇总一个作品实际占用的本地 / 云端文件引用。引用值统一为
// localartifact 的文件名，因此既能删本地缓存，也能通过 media_objects 找到
// 对应的 TOS object_key；合并成片直接上传 TOS 时则单独返回 object_key。
func (s *MySQLStore) bookStorageRefs(ctx context.Context, owner, bookID string) (map[string]struct{}, map[string]struct{}, error) {
	refs := map[string]struct{}{}
	tosKeys := map[string]struct{}{}
	addRef := func(value string) {
		if value = strings.TrimSpace(value); value != "" {
			refs[value] = struct{}{}
		}
	}
	addKey := func(value string) {
		if value = strings.TrimSpace(value); value != "" {
			tosKeys[value] = struct{}{}
		}
	}
	queries := []struct {
		query string
		args  []any
	}{
		{`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id WHERE p.owner_username=? AND p.book_id=?`, []any{owner, bookID}},
		{`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id WHERE a.owner_username=? AND a.book_id=?`, []any{owner, bookID}},
		{`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_collection_items i JOIN batch_factory_v12_comic_collections c ON c.id=i.collection_id WHERE c.owner_username=? AND c.book_id=?`, []any{owner, bookID}},
	}
	for _, item := range queries {
		rows, err := s.db.QueryContext(ctx, item.query, item.args...)
		if err != nil {
			return nil, nil, err
		}
		for rows.Next() {
			var ref string
			if err := rows.Scan(&ref); err != nil {
				rows.Close()
				return nil, nil, err
			}
			addRef(ref)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, nil, err
		}
		rows.Close()
	}
	mediaRefs, artifactIDs, mediaKeys, err := s.bookMediaURLs(ctx, `SELECT COALESCE(j.output_url,'') FROM batch_factory_v12_merge_jobs j WHERE j.owner_username=? AND j.book_id=?`, owner, bookID)
	if err != nil {
		return nil, nil, err
	}
	for _, ref := range mediaRefs {
		addRef(ref)
	}
	for _, key := range mediaKeys {
		addKey(key)
	}
	for _, query := range []string{
		`SELECT COALESCE(t.media_url,'') FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id WHERE j.owner_username=? AND j.book_id=?`,
		`SELECT COALESCE(s.media_url,'') FROM batch_factory_v12_merge_sources s JOIN batch_factory_v12_merge_jobs m ON m.id=s.job_id WHERE s.owner_username=? AND m.book_id=?`,
	} {
		moreRefs, moreIDs, moreKeys, err := s.bookMediaURLs(ctx, query, owner, bookID)
		if err != nil {
			return nil, nil, err
		}
		for _, ref := range moreRefs {
			addRef(ref)
		}
		for _, key := range moreKeys {
			addKey(key)
		}
		artifactIDs = append(artifactIDs, moreIDs...)
	}
	artifactRefs, err := s.artifactStorageRefs(ctx, owner, artifactIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, ref := range artifactRefs {
		addRef(ref)
	}
	return refs, tosKeys, nil
}

// visibleBookStorageRefs 返回其他可见作品仍在使用、因此不能删除的引用。
func (s *MySQLStore) visibleBookStorageRefs(ctx context.Context, owner, excludeBookID string) (map[string]struct{}, error) {
	refs := map[string]struct{}{}
	addRef := func(value string) {
		if value = strings.TrimSpace(value); value != "" {
			refs[value] = struct{}{}
		}
	}
	for _, query := range []string{
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id JOIN batch_factory_v12_books b ON b.id=p.book_id AND b.owner_username=p.owner_username WHERE i.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id JOIN batch_factory_v12_books b ON b.id=a.book_id AND b.owner_username=a.owner_username WHERE a.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_collection_items i JOIN batch_factory_v12_comic_collections c ON c.id=i.collection_id JOIN batch_factory_v12_books b ON b.id=c.book_id AND b.owner_username=c.owner_username WHERE c.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`,
	} {
		rows, err := s.db.QueryContext(ctx, query, owner, excludeBookID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				rows.Close()
				return nil, err
			}
			addRef(value)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	mergeRefs, _, _, err := s.bookMediaURLs(ctx, `SELECT COALESCE(j.output_url,'') FROM batch_factory_v12_merge_jobs j JOIN batch_factory_v12_books b ON b.id=j.book_id AND b.owner_username=j.owner_username WHERE j.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`, owner, excludeBookID)
	if err != nil {
		return nil, err
	}
	for _, ref := range mergeRefs {
		addRef(ref)
	}
	artifactIDs := []string{}
	for _, query := range []string{
		`SELECT COALESCE(t.media_url,'') FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id JOIN batch_factory_v12_books b ON b.id=j.book_id AND b.owner_username=j.owner_username WHERE j.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`,
		`SELECT COALESCE(s.media_url,'') FROM batch_factory_v12_merge_sources s JOIN batch_factory_v12_merge_jobs m ON m.id=s.job_id JOIN batch_factory_v12_books b ON b.id=m.book_id AND b.owner_username=m.owner_username WHERE s.owner_username=? AND b.id<>? AND b.hidden_at IS NULL`,
	} {
		rows, err := s.db.QueryContext(ctx, query, owner, excludeBookID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				rows.Close()
				return nil, err
			}
			if id, ok := parseLocalExecutorArtifactID(value); ok {
				artifactIDs = append(artifactIDs, id)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	artifactRefs, err := s.artifactStorageRefs(ctx, owner, artifactIDs)
	if err != nil {
		return nil, err
	}
	for _, ref := range artifactRefs {
		addRef(ref)
	}
	return refs, nil
}

// bookMediaURLs 把媒体 URL 拆成“本地合并视频文件名”“本地执行器 artifact id”
// 和“直接上传 TOS 的合并成片 object_key”。
func (s *MySQLStore) bookMediaURLs(ctx context.Context, query string, args ...any) ([]string, []string, []string, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	refs := []string{}
	ids := []string{}
	keys := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, nil, nil, err
		}
		if ref, ok := mergeMediaStorageRef(value); ok {
			refs = append(refs, ref)
			continue
		}
		if key, ok := mergeTOSObjectKey(value); ok {
			keys = append(keys, key)
			continue
		}
		if id, ok := parseLocalExecutorArtifactID(value); ok {
			ids = append(ids, id)
		}
	}
	return refs, ids, keys, rows.Err()
}

func (s *MySQLStore) artifactStorageRefs(ctx context.Context, owner string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	unique := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	placeholders := make([]string, 0, len(unique))
	args := make([]any, 0, len(unique)+1)
	args = append(args, owner)
	for _, id := range unique {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(storage_ref,'') FROM local_executor_artifacts WHERE owner_username=? AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := []string{}
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, err
		}
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs, rows.Err()
}

// deleteRetentionBookRows 删除作品数据库记录。若批次内已无作品，同时清掉批次
// 记录，避免列表里留下永远空的项目。
func (s *MySQLStore) deleteRetentionBookRows(ctx context.Context, book RetentionBook) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var hidden sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT hidden_at FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, book.ID, book.BatchID, book.Owner).Scan(&hidden)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	// 缓冲期内用户调大保留天数时作品会被恢复显示，此时必须放弃删除。
	if !hidden.Valid {
		return tx.Commit()
	}
	// 合并任务只按批次关联，不会随作品级联删除。先清掉该作品的合并记录，
	// 避免项目卡片继续引用已经删除的成片文件。
	if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v12_merge_sources WHERE owner_username=? AND job_id IN (SELECT id FROM batch_factory_v12_merge_jobs WHERE owner_username=? AND batch_id=? AND book_id=?)`, book.Owner, book.Owner, book.BatchID, book.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v12_merge_jobs WHERE owner_username=? AND batch_id=? AND book_id=?`, book.Owner, book.BatchID, book.ID); err != nil {
		return err
	}
	if err := deleteBookTx(ctx, tx, book.Owner, book.BatchID, book.ID); err != nil {
		return err
	}
	var remaining int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=?`, book.BatchID, book.Owner).Scan(&remaining); err != nil {
		return err
	}
	if remaining == 0 {
		if err := deleteBatchTx(ctx, tx, book.Owner, book.BatchID); err != nil {
			return err
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET revision=revision+1,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner_username=?`, book.BatchID, book.Owner); err != nil {
		return err
	}
	return tx.Commit()
}

// mergeMediaStorageRef 从合并视频 URL 中提取文件名（例如 merge_xxx.mp4）。
func mergeMediaStorageRef(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	if !strings.Contains(parsed.Path, "/merge-media/") {
		return "", false
	}
	value := path.Base(parsed.Path)
	if !strings.HasPrefix(value, "merge_") {
		return "", false
	}
	if !strings.HasSuffix(value, ".mp4") {
		value += ".mp4"
	}
	if !localMergedArtifactID.MatchString(strings.TrimSuffix(value, ".mp4")) {
		return "", false
	}
	return value, true
}

// mergeTOSObjectKey 从直接上传 TOS 的合并成片 / 上传视频 URL 中还原 object_key。
func mergeTOSObjectKey(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	for _, prefix := range []string{"batch-merged", "batch-uploaded"} {
		if !strings.Contains(parsed.Path, "/"+prefix+"/") {
			continue
		}
		name := path.Base(parsed.Path)
		if name == "" || name == "." || name == "/" {
			return "", false
		}
		return prefix + "/" + name, true
	}
	return "", false
}

// parseLocalExecutorArtifactID 从本地执行器媒体 URL 中提取 artifact id。
func parseLocalExecutorArtifactID(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	marker := "/api/local-executor/v1/artifacts/"
	if !strings.Contains(parsed.Path+"/", marker) {
		return "", false
	}
	rest := parsed.Path[strings.Index(parsed.Path, marker)+len(marker):]
	id := strings.TrimSuffix(path.Base(rest), "/")
	if id == "" || id == "." {
		return "", false
	}
	return id, true
}
