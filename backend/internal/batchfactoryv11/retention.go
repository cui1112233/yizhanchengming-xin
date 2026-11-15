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
	// DefaultRetentionGrace 是作品集到期隐藏后到真正删除之间的缓冲时间。
	DefaultRetentionGrace = 24 * time.Hour
	retentionClaimWindow  = 15 * time.Minute
	retentionMaxPerRun    = 50
)

// RemoteObjectDeleter 只暴露删除单个对象的能力，避免清理器拿到整份 TOS 客户端。
type RemoteObjectDeleter interface {
	DeleteObject(context.Context, string) error
}

// RetentionBatch 是一条待硬删的作品集（批次）记录。
type RetentionBatch struct {
	ID    string
	Owner string
}

// RetentionRecomputeSummary 描述一次重算的隐藏 / 恢复数量。
type RetentionRecomputeSummary struct {
	Hidden   int64
	Restored int64
}

// BatchRetentionSupervisor 按作品集维度执行保留策略：重算到期时间，并把到期
// 作品集移入归档（隐藏）。归档作品永久保留，只有用户手动彻底删除才释放文件。
type BatchRetentionSupervisor struct {
	Store         *MySQLStore
	Local         *localartifact.Store
	Remote        RemoteObjectDeleter
	RetentionDays func(string) int
	MaxPerRun     int
}

// RunOnce 执行一次完整巡检。到期作品只归档，不自动删除。
func (s BatchRetentionSupervisor) RunOnce(ctx context.Context) error {
	if s.Store == nil {
		return fmt.Errorf("batch retention store is not configured")
	}
	owners, err := s.Store.ListRetentionOwners(ctx)
	if err != nil {
		return err
	}
	for _, owner := range owners {
		days := DefaultLocalMergedArtifactRetentionDays
		if s.RetentionDays != nil {
			days = normalizeLocalMergedArtifactRetentionDays(s.RetentionDays(owner))
		}
		if _, err := s.Store.RecomputeBatchRetention(ctx, owner, days); err != nil {
			return err
		}
	}
	return nil
}

// StartDaily 每天在指定本地时间执行一次巡检。归档动作集中在凌晨，避免白天
// 业务高峰期间批量修改作品可见状态。
func (s BatchRetentionSupervisor) StartDaily(hour, minute int) {
	go func() {
		for {
			next := nextRetentionSweep(hour, minute, time.Now())
			log.Printf("batch retention next sweep at %s", next.Format("2006-01-02 15:04:05"))
			timer := time.NewTimer(time.Until(next))
			<-timer.C
			if err := s.RunOnce(context.Background()); err != nil {
				log.Printf("batch retention sweep failed: %v", err)
			}
		}
	}()
}

// nextRetentionSweep 返回下一次巡检时间；当天时间已过就顺延到第二天。
func nextRetentionSweep(hour, minute int, now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// ListRetentionOwners 返回仍有作品集的账号，避免对空账号做无意义扫描。
func (s *MySQLStore) ListRetentionOwners(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT owner_username FROM batch_factory_v12_batches`)
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

type retentionBatchActivity struct {
	id          string
	completedAt time.Time
	existing    sql.NullTime
}

// RecomputeBatchRetention 按当前保留天数重算该账号所有作品集的到期时间。
// 它同时处理三件事：恢复旧版本按单本书隐藏的数据；把未到期但已隐藏的作品集
// 恢复显示；把已到期的作品集隐藏并安排 24 小时后的硬删。每次巡检都会重算，
// 因此修改天数会追溯已有作品集。
func (s *MySQLStore) RecomputeBatchRetention(ctx context.Context, owner string, days int) (RetentionRecomputeSummary, error) {
	summary := RetentionRecomputeSummary{}
	if strings.TrimSpace(owner) == "" {
		return summary, ErrInvalid
	}
	days = normalizeLocalMergedArtifactRetentionDays(days)
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,
GREATEST(
  b.updated_at,
  COALESCE((SELECT MAX(x.updated_at) FROM batch_factory_v12_books x WHERE x.batch_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(s.updated_at) FROM batch_factory_v12_book_stage_runs s JOIN batch_factory_v12_books x ON x.id=s.book_id WHERE x.batch_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(m.updated_at) FROM batch_factory_v12_merge_jobs m WHERE m.batch_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(p.updated_at) FROM batch_factory_v12_comic_pages p JOIN batch_factory_v12_books x ON x.id=p.book_id WHERE x.batch_id=b.id), b.updated_at),
  COALESCE((SELECT MAX(j.updated_at) FROM batch_factory_v12_production_jobs j WHERE j.batch_id=b.id), b.updated_at)
) AS completed_at,
b.completed_at
FROM batch_factory_v12_batches b WHERE b.owner_username=?`, owner)
	if err != nil {
		return summary, err
	}
	activities := []retentionBatchActivity{}
	for rows.Next() {
		var item retentionBatchActivity
		if err := rows.Scan(&item.id, &item.completedAt, &item.existing); err != nil {
			rows.Close()
			return summary, err
		}
		// 完成时间只前进不后退：删除分镜或重做合成都不应让作品集“变新”。
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
	// 旧版本按单本书隐藏的数据恢复为可见，改用作品集统一计时后不会留下
	// 半隐藏的批次。
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET hidden_at=NULL,purge_after=NULL WHERE owner_username=? AND hidden_at IS NOT NULL`, owner); err != nil {
		return summary, err
	}
	for _, item := range activities {
		expires := retentionExpiry(item.completedAt, days)
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET completed_at=?,retention_expires_at=? WHERE id=? AND owner_username=?`, item.completedAt, expires, item.id, owner); err != nil {
			return summary, err
		}
	}
	// 调大保留天数后，未到期的归档作品集自动回到列表。
	if result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET hidden_at=NULL,purge_after=NULL WHERE owner_username=? AND hidden_at IS NOT NULL AND retention_expires_at>?`, owner, now); err != nil {
		return summary, err
	} else if affected, err := result.RowsAffected(); err == nil {
		summary.Restored = affected
	}
	// 到期后只进入归档：永久保留，等待用户恢复或手动彻底删除。
	if result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET hidden_at=?,purge_after=NULL WHERE owner_username=? AND hidden_at IS NULL AND retention_expires_at IS NOT NULL AND retention_expires_at<=?`, now, owner, now); err != nil {
		return summary, err
	} else if affected, err := result.RowsAffected(); err == nil {
		summary.Hidden = affected
	}
	if err := tx.Commit(); err != nil {
		return summary, err
	}
	return summary, nil
}

// retentionExpiry 把完成时间换算成“完成日 + N 天”的本地零点。例如 11-10 完成、
// 保留 7 天，会在 11-17 00:00 的巡检中归档。
func retentionExpiry(completed time.Time, days int) time.Time {
	if days <= 0 {
		days = DefaultLocalMergedArtifactRetentionDays
	}
	local := completed.In(time.Local)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local)
	return start.AddDate(0, 0, days).UTC()
}

// BatchRetentionInfo 是作品集归档库需要的展示字段。
type BatchRetentionInfo struct {
	ID          string
	Title       string
	CompletedAt time.Time
	ExpiresAt   sql.NullTime
	HiddenAt    sql.NullTime
}

// ListBatchRetention 返回账号下所有作品集的保留状态（含已归档）。
func (s *MySQLStore) ListBatchRetention(ctx context.Context, owner string) ([]BatchRetentionInfo, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id, COALESCE(NULLIF(r.title,''),'未命名作品集'),
COALESCE(b.completed_at,b.updated_at), b.retention_expires_at, b.hidden_at
FROM batch_factory_v12_batches b
LEFT JOIN batch_factory_v12_batch_records r ON r.batch_id=b.id
WHERE b.owner_username=? ORDER BY b.created_at DESC,b.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BatchRetentionInfo{}
	for rows.Next() {
		var item BatchRetentionInfo
		if err := rows.Scan(&item.ID, &item.Title, &item.CompletedAt, &item.ExpiresAt, &item.HiddenAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListDueArchivedBatches 返回归档时间已超过可恢复期的作品集。
func (s *MySQLStore) ListDueArchivedBatches(ctx context.Context, owner string, cutoff time.Time) ([]RetentionBatch, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,owner_username FROM batch_factory_v12_batches WHERE owner_username=? AND hidden_at IS NOT NULL AND hidden_at<=? ORDER BY hidden_at,id LIMIT 50`, owner, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RetentionBatch{}
	for rows.Next() {
		var batch RetentionBatch
		if err := rows.Scan(&batch.ID, &batch.Owner); err != nil {
			return nil, err
		}
		out = append(out, batch)
	}
	return out, rows.Err()
}

// RestoreArchivedBatch 恢复归档作品集，并按当前保留天数从恢复时刻重新计时。
func (s *MySQLStore) RestoreArchivedBatch(ctx context.Context, owner, batchID string, days int, now time.Time) error {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return ErrInvalid
	}
	days = normalizeLocalMergedArtifactRetentionDays(days)
	expires := now.UTC().AddDate(0, 0, days)
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET hidden_at=NULL,purge_after=NULL,completed_at=?,retention_expires_at=? WHERE id=? AND owner_username=? AND hidden_at IS NOT NULL`, now.UTC(), expires, batchID, owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrNotFound
	}
	_, err = s.db.ExecContext(ctx, `UPDATE batch_factory_v12_books SET hidden_at=NULL,purge_after=NULL WHERE batch_id=? AND owner_username=?`, batchID, owner)
	return err
}

// PurgeArchivedBatch 彻底删除已归档作品集：先删 TOS/本地文件，再删数据库记录。
func (s *MySQLStore) PurgeArchivedBatch(ctx context.Context, owner, batchID string, local *localartifact.Store, remote RemoteObjectDeleter) error {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return ErrInvalid
	}
	var hidden sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT hidden_at FROM batch_factory_v12_batches WHERE id=? AND owner_username=?`, batchID, owner).Scan(&hidden)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !hidden.Valid {
		return ErrInvalid
	}
	return s.PurgeRetentionBatch(ctx, RetentionBatch{ID: batchID, Owner: owner}, local, remote)
}

// ClaimDueRetentionBatch 原子领取一个已过缓冲期的作品集。purge_after 会被推后，
// 这样进程在删除中途退出时不会立即被下一次巡检重复抢占。
func (s *MySQLStore) ClaimDueRetentionBatch(ctx context.Context) (RetentionBatch, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RetentionBatch{}, false, err
	}
	defer tx.Rollback()
	var batch RetentionBatch
	// 清理时间统一按 UTC 写入，因此这里也必须用 UTC_TIMESTAMP 比较；数据库
	// 会话时区是本地时间时，CURRENT_TIMESTAMP 会提前 8 小时触发删除。
	err = tx.QueryRowContext(ctx, `SELECT id,owner_username FROM batch_factory_v12_batches WHERE hidden_at IS NOT NULL AND purge_after<=UTC_TIMESTAMP(6) ORDER BY purge_after,id LIMIT 1 FOR UPDATE SKIP LOCKED`).
		Scan(&batch.ID, &batch.Owner)
	if errors.Is(err, sql.ErrNoRows) {
		if commitErr := tx.Commit(); commitErr != nil {
			return RetentionBatch{}, false, commitErr
		}
		return RetentionBatch{}, false, nil
	}
	if err != nil {
		return RetentionBatch{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET purge_after=? WHERE id=?`, time.Now().UTC().Add(retentionClaimWindow), batch.ID); err != nil {
		return RetentionBatch{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return RetentionBatch{}, false, err
	}
	return batch, true, nil
}

// PurgeRetentionBatch 先删云端与本地文件，全部成功后再删数据库记录。任何一步
// 失败都会保留记录并等待下一轮重试，避免出现“记录没了、文件还在”的孤儿空间。
func (s *MySQLStore) PurgeRetentionBatch(ctx context.Context, batch RetentionBatch, local *localartifact.Store, remote RemoteObjectDeleter) error {
	if strings.TrimSpace(batch.Owner) == "" || strings.TrimSpace(batch.ID) == "" {
		return ErrInvalid
	}
	if local == nil {
		return fmt.Errorf("batch retention local store is not configured")
	}
	refs, tosKeys, err := s.batchStorageRefs(ctx, batch.Owner, batch.ID)
	if err != nil {
		return err
	}
	// 没有 TOS 客户端时不能确定云端对象是否已删除，保留作品集等待下次重试，
	// 避免出现“数据库删了、云端还在”的孤儿文件。
	if remote == nil && len(tosKeys) > 0 {
		return fmt.Errorf("TOS is not configured; retain batch and cloud references for retry")
	}
	if len(refs) > 0 {
		shared, err := s.visibleBatchStorageRefs(ctx, batch.Owner, batch.ID)
		if err != nil {
			return err
		}
		for ref := range refs {
			if _, ok := shared[ref]; ok {
				continue
			}
			if err := s.purgeStorageRef(ctx, batch.Owner, ref, local, remote); err != nil {
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
	return s.deleteRetentionBatchRows(ctx, batch)
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

// batchStorageRefs 汇总一个作品集实际占用的本地 / 云端文件引用。引用值统一为
// localartifact 的文件名，因此既能删本地缓存，也能通过 media_objects 找到
// 对应的 TOS object_key；合并成片直接上传 TOS 时则单独返回 object_key。
func (s *MySQLStore) batchStorageRefs(ctx context.Context, owner, batchID string) (map[string]struct{}, map[string]struct{}, error) {
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
	queries := []string{
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id JOIN batch_factory_v12_books b ON b.id=p.book_id WHERE p.owner_username=? AND b.batch_id=?`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id JOIN batch_factory_v12_books b ON b.id=a.book_id WHERE a.owner_username=? AND b.batch_id=?`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_collection_items i JOIN batch_factory_v12_comic_collections c ON c.id=i.collection_id JOIN batch_factory_v12_books b ON b.id=c.book_id WHERE c.owner_username=? AND b.batch_id=?`,
	}
	for _, query := range queries {
		rows, err := s.db.QueryContext(ctx, query, owner, batchID)
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
	mediaRefs, artifactIDs, mediaKeys, err := s.batchMediaURLs(ctx, `SELECT COALESCE(m.output_url,'') FROM batch_factory_v12_merge_jobs m WHERE m.owner_username=? AND m.batch_id=?`, owner, batchID)
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
		`SELECT COALESCE(t.media_url,'') FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id WHERE j.owner_username=? AND j.batch_id=?`,
		`SELECT COALESCE(s.media_url,'') FROM batch_factory_v12_merge_sources s JOIN batch_factory_v12_merge_jobs m ON m.id=s.job_id WHERE s.owner_username=? AND m.batch_id=?`,
	} {
		moreRefs, moreIDs, moreKeys, err := s.batchMediaURLs(ctx, query, owner, batchID)
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

// visibleBatchStorageRefs 返回其他可见作品集仍在使用、因此不能删除的引用。
func (s *MySQLStore) visibleBatchStorageRefs(ctx context.Context, owner, excludeBatchID string) (map[string]struct{}, error) {
	refs := map[string]struct{}{}
	addRef := func(value string) {
		if value = strings.TrimSpace(value); value != "" {
			refs[value] = struct{}{}
		}
	}
	for _, query := range []string{
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id JOIN batch_factory_v12_books b ON b.id=p.book_id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id WHERE i.owner_username=? AND b.batch_id<>? AND ba.hidden_at IS NULL`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id JOIN batch_factory_v12_books b ON b.id=a.book_id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id WHERE a.owner_username=? AND b.batch_id<>? AND ba.hidden_at IS NULL`,
		`SELECT COALESCE(i.storage_ref,'') FROM batch_factory_v12_comic_collection_items i JOIN batch_factory_v12_comic_collections c ON c.id=i.collection_id JOIN batch_factory_v12_books b ON b.id=c.book_id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id WHERE c.owner_username=? AND b.batch_id<>? AND ba.hidden_at IS NULL`,
	} {
		rows, err := s.db.QueryContext(ctx, query, owner, excludeBatchID)
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
	mergeRefs, _, _, err := s.batchMediaURLs(ctx, `SELECT COALESCE(m.output_url,'') FROM batch_factory_v12_merge_jobs m JOIN batch_factory_v12_batches b ON b.id=m.batch_id AND b.owner_username=m.owner_username WHERE m.owner_username=? AND m.batch_id<>? AND b.hidden_at IS NULL`, owner, excludeBatchID)
	if err != nil {
		return nil, err
	}
	for _, ref := range mergeRefs {
		addRef(ref)
	}
	artifactIDs := []string{}
	for _, query := range []string{
		`SELECT COALESCE(t.media_url,'') FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id JOIN batch_factory_v12_batches b ON b.id=j.batch_id AND b.owner_username=j.owner_username WHERE j.owner_username=? AND j.batch_id<>? AND b.hidden_at IS NULL`,
		`SELECT COALESCE(s.media_url,'') FROM batch_factory_v12_merge_sources s JOIN batch_factory_v12_merge_jobs m ON m.id=s.job_id JOIN batch_factory_v12_batches b ON b.id=m.batch_id AND b.owner_username=m.owner_username WHERE s.owner_username=? AND m.batch_id<>? AND b.hidden_at IS NULL`,
	} {
		rows, err := s.db.QueryContext(ctx, query, owner, excludeBatchID)
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

// batchMediaURLs 把媒体 URL 拆成“本地合并视频文件名”“本地执行器 artifact id”
// 和“直接上传 TOS 的合并成片 object_key”。
func (s *MySQLStore) batchMediaURLs(ctx context.Context, query string, args ...any) ([]string, []string, []string, error) {
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

// deleteRetentionBatchRows 删除整个作品集的数据库记录：先删每本书的子表，再删
// 批次本身。缓冲期内用户调大保留天数时作品集会恢复显示，此时必须放弃删除。
func (s *MySQLStore) deleteRetentionBatchRows(ctx context.Context, batch RetentionBatch) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var hidden sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT hidden_at FROM batch_factory_v12_batches WHERE id=? AND owner_username=? FOR UPDATE`, batch.ID, batch.Owner).Scan(&hidden)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if !hidden.Valid {
		return tx.Commit()
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=?`, batch.ID, batch.Owner)
	if err != nil {
		return err
	}
	bookIDs := []string{}
	for rows.Next() {
		var bookID string
		if err := rows.Scan(&bookID); err != nil {
			rows.Close()
			return err
		}
		bookIDs = append(bookIDs, bookID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, bookID := range bookIDs {
		if err := deleteBookTx(ctx, tx, batch.Owner, batch.ID, bookID); err != nil {
			return err
		}
	}
	if err := deleteBatchTx(ctx, tx, batch.Owner, batch.ID); err != nil {
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
