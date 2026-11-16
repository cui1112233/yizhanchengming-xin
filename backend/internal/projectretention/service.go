// Package projectretention 统一管理“作品保留时长 → 到期归档 → 恢复/彻底删除”的
// 全站生命周期。批量工厂沿用批次表上的保留字段，其余模块统一记录在
// project_retention 表中。
package projectretention

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

type Module string

const (
	// ModuleBatchFactory 表示批量工厂的作品集（批次）。
	ModuleBatchFactory Module = "batch_factory"
	// ModuleComicScreenplay 表示漫剧创作的作品集（项目）。
	ModuleComicScreenplay Module = "comic_screenplay"
	// ModuleCMCanvas 表示 CM 画布的作品集（项目）。
	ModuleCMCanvas Module = "cm_canvas"
	// ModuleNovelCreation 表示小说创作的作品集（项目）。
	ModuleNovelCreation Module = "novel_creation"

	// DefaultRetentionDays 与账号设置 productionRetentionDays 的默认值保持一致。
	DefaultRetentionDays = 7
	// DefaultArchiveGrace 是作品进入归档库后、被彻底删除前的可恢复时间。
	DefaultArchiveGrace = 24 * time.Hour
)

var (
	ErrPurging  = errors.New("作品正在清理，无法恢复；清理失败可重试删除")
	ErrInvalid  = errors.New("保留/归档参数无效")
	ErrNotFound = errors.New("作品不存在或未归档")
)

// Record 是列表页与归档库共用的保留状态。
type Record struct {
	Purging       bool       `json:"purging,omitempty"`
	Module        Module     `json:"module"`
	ProjectID     string     `json:"projectId"`
	Title         string     `json:"title"`
	LastActivity  time.Time  `json:"lastActivityAt"`
	ExpiresAt     time.Time  `json:"retentionExpiresAt"`
	ArchivedAt    *time.Time `json:"archivedAt,omitempty"`
	RemainingDays int        `json:"remainingDays"`
}

// Service 负责每日 00:00 重算保留期、到期归档，以及归档作品的恢复与彻底删除。
type Service struct {
	NovelObjects  func(context.Context, string, string) error
	DB            *sql.DB
	Batch         *batchfactoryv11.MySQLStore
	Local         *localartifact.Store
	Remote        batchfactoryv11.RemoteObjectDeleter
	CanvasAssets  agentworkspace.AssetStore
	CanvasObjects agentworkspace.AssetObjectStore
	Days          func(owner string) int
	Now           func() time.Time
	Location      *time.Location
	// Grace 控制归档后可恢复的时长，默认 24 小时。
	Grace time.Duration
}

// NewService 构造保留服务。batch 可为空，表示当前运行环境未启用批量工厂存储。
func NewService(db *sql.DB, batch *batchfactoryv11.MySQLStore, local *localartifact.Store, remote batchfactoryv11.RemoteObjectDeleter, days func(string) int) *Service {
	return &Service{DB: db, Batch: batch, Local: local, Remote: remote, Days: days, Now: time.Now, Location: time.Local}
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) loc() *time.Location {
	if s.Location != nil {
		return s.Location
	}
	return time.Local
}

func (s *Service) days(owner string) int {
	days := DefaultRetentionDays
	if s.Days != nil {
		days = s.Days(owner)
	}
	if days <= 0 {
		days = DefaultRetentionDays
	}
	return days
}

// expiresAt 把最后一次活动时间换算成“最后活动日 + N 天”的本地零点。
// 例如 11-10 任意时刻完成、保留 7 天，会在 11-17 00:00 归档。
func (s *Service) expiresAt(last time.Time, days int) time.Time {
	loc := s.loc()
	local := last.In(loc)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return start.AddDate(0, 0, days).UTC()
}

func remainingDays(expires, now time.Time) int {
	if expires.IsZero() {
		return 0
	}
	days := int(math.Ceil(expires.Sub(now).Hours() / 24))
	if days < 0 {
		return 0
	}
	return days
}

// StartDaily 每天在指定本地时间执行一次巡检。到期作品只会进入归档，不会被自动删除。
func (s *Service) StartDaily(hour, minute int) {
	go func() {
		for {
			next := nextSweep(hour, minute, time.Now(), s.loc())
			log.Printf("project retention next sweep at %s", next.Format("2006-01-02 15:04:05"))
			timer := time.NewTimer(time.Until(next))
			<-timer.C
			if err := s.RunOnce(context.Background()); err != nil {
				log.Printf("project retention sweep failed: %v", err)
			}
		}
	}()
}

func nextSweep(hour, minute int, now time.Time, loc *time.Location) time.Time {
	local := now.In(loc)
	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// RunOnce 对每个有作品的账号重算保留期、归档到期作品，并清理超过恢复期的归档。
func (s *Service) RunOnce(ctx context.Context) error {
	if s.DB == nil {
		return fmt.Errorf("project retention database is not configured")
	}
	owners, err := s.owners(ctx)
	if err != nil {
		return err
	}
	for _, owner := range owners {
		if err := s.recompute(ctx, owner); err != nil {
			log.Printf("project retention recompute failed owner=%s: %v", owner, err)
		}
		if err := s.purgeExpired(ctx, owner); err != nil {
			log.Printf("project retention purge failed owner=%s: %v", owner, err)
		}
	}
	return nil
}

// purgeExpired 彻底删除已归档超过 Grace 的作品集，释放数据库、TOS 与本地文件。
func (s *Service) purgeExpired(ctx context.Context, owner string) error {
	grace := s.Grace
	if grace <= 0 {
		grace = DefaultArchiveGrace
	}
	cutoff := s.now().UTC().Add(-grace)
	if s.Batch != nil {
		batches, err := s.Batch.ListDueArchivedBatches(ctx, owner, cutoff)
		if err != nil {
			return err
		}
		for _, batch := range batches {
			if err := s.Batch.PurgeArchivedBatch(ctx, owner, batch.ID, s.Local, s.Remote); err != nil {
				log.Printf("batch retention purge failed owner=%s batch=%s: %v", owner, batch.ID, err)
			}
		}
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT module, project_id FROM project_retention WHERE owner_username=? AND archived_at IS NOT NULL AND archived_at<=?`, owner, cutoff)
	if err != nil {
		return err
	}
	type due struct {
		module    Module
		projectID string
	}
	items := []due{}
	for rows.Next() {
		var module string
		var projectID string
		if err := rows.Scan(&module, &projectID); err != nil {
			rows.Close()
			return err
		}
		items = append(items, due{module: Module(module), projectID: projectID})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, item := range items {
		if err := s.Purge(ctx, owner, item.module, item.projectID); err != nil {
			log.Printf("project retention purge failed owner=%s module=%s project=%s: %v", owner, item.module, item.projectID, err)
		}
	}
	return nil
}

func (s *Service) owners(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT owner_username FROM screenplay_projects WHERE deleted_at IS NULL
UNION SELECT owner_username FROM agent_projects WHERE deleted_at IS NULL
UNION SELECT owner_username FROM novel_projects WHERE deleted_at IS NULL
UNION SELECT owner_username FROM batch_factory_v12_batches`)
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

type moduleQuery struct {
	module Module
	sql    string
	// extraSQL 用于补充跨表活动时间。部分历史表与项目表的字符集/排序规则
	// 不一致，直接 JOIN 会触发 Illegal mix of collations，因此单独查询后在
	// Go 侧取最大值。
	extraSQL string
	// validSQL 返回该模块仍然存在的项目 ID，用于清理已删除作品的保留记录。
	validSQL string
}

var moduleQueries = []moduleQuery{
	{
		module: ModuleComicScreenplay,
		sql: `SELECT p.id, COALESCE(NULLIF(p.title,''),'未命名作品集'), GREATEST(
  p.updated_at,
  COALESCE((SELECT MAX(d.updated_at) FROM screenplay_documents d WHERE d.owner_username=p.owner_username AND d.project_id=p.id), p.updated_at),
  COALESCE((SELECT MAX(a.updated_at) FROM screenplay_assets a WHERE a.owner_username=p.owner_username AND a.project_id=p.id), p.updated_at),
  COALESCE((SELECT MAX(s.updated_at) FROM screenplay_storyboards s WHERE s.owner_username=p.owner_username AND s.project_id=p.id), p.updated_at)
) FROM screenplay_projects p WHERE p.owner_username=? AND p.deleted_at IS NULL`,
		extraSQL: `SELECT project_id, MAX(updated_at) FROM screenplay_production_jobs WHERE owner_username=? GROUP BY project_id`,
		validSQL: `SELECT id FROM screenplay_projects WHERE owner_username=? AND deleted_at IS NULL`,
	},
	{
		module: ModuleCMCanvas,
		sql: `SELECT p.id, COALESCE(NULLIF(p.title,''),'未命名作品集'), GREATEST(
  p.updated_at,
  COALESCE((SELECT MAX(c.updated_at) FROM agent_canvases c WHERE c.owner_username=p.owner_username AND c.project_id=p.id), p.updated_at),
  COALESCE((SELECT MAX(n.updated_at) FROM agent_canvas_nodes n WHERE n.owner_username=p.owner_username AND n.project_id=p.id), p.updated_at),
  COALESCE((SELECT MAX(e.updated_at) FROM agent_executions e WHERE e.owner_username=p.owner_username AND e.project_id=p.id), p.updated_at)
) FROM agent_projects p WHERE p.owner_username=? AND p.deleted_at IS NULL`,
		validSQL: `SELECT id FROM agent_projects WHERE owner_username=? AND deleted_at IS NULL`,
	},
	{
		module:   ModuleNovelCreation,
		sql:      novelActivitySQL,
		validSQL: `SELECT project_id FROM novel_projects WHERE owner_username=? AND deleted_at IS NULL`,
	},
}

func (s *Service) recompute(ctx context.Context, owner string) error {
	days := s.days(owner)
	now := s.now().UTC()
	for _, query := range moduleQueries {
		if err := s.recomputeModule(ctx, owner, days, now, query); err != nil {
			return err
		}
	}
	if s.Batch != nil {
		if _, err := s.Batch.RecomputeBatchRetention(ctx, owner, days); err != nil {
			return err
		}
	}
	return s.archiveDue(ctx, owner)
}

func (s *Service) recomputeModule(ctx context.Context, owner string, days int, now time.Time, query moduleQuery) error {
	rows, err := s.DB.QueryContext(ctx, query.sql, owner)
	if err != nil {
		return err
	}
	type activity struct {
		id    string
		title string
		last  time.Time
	}
	items := []activity{}
	for rows.Next() {
		var item activity
		if err := rows.Scan(&item.id, &item.title, &item.last); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	if query.extraSQL != "" {
		extra, err := s.activityByProject(ctx, query.extraSQL, owner)
		if err != nil {
			return err
		}
		for i := range items {
			if last, ok := extra[items[i].id]; ok && last.After(items[i].last) {
				items[i].last = last
			}
		}
	}

	valid, err := s.validProjectIDs(ctx, query.validSQL, owner)
	if err != nil {
		return err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		if item.last.IsZero() {
			continue
		}
		expires := s.expiresAt(item.last, days)
		if _, err := tx.ExecContext(ctx, `INSERT INTO project_retention
(owner_username, module, project_id, title, last_activity_at, retention_expires_at, archived_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,NULL,?,?)
ON DUPLICATE KEY UPDATE title=VALUES(title), last_activity_at=VALUES(last_activity_at), retention_expires_at=VALUES(retention_expires_at), updated_at=VALUES(updated_at)`,
			owner, string(query.module), item.id, item.title, item.last.UTC(), expires, now, now); err != nil {
			return err
		}
	}
	stale, err := s.staleProjectIDs(ctx, owner, query.module, valid)
	if err != nil {
		return err
	}
	for _, projectID := range stale {
		if _, err := tx.ExecContext(ctx, `DELETE FROM project_retention WHERE owner_username=? AND module=? AND project_id=?`, owner, string(query.module), projectID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) activityByProject(ctx context.Context, querySQL, owner string) (map[string]time.Time, error) {
	rows, err := s.DB.QueryContext(ctx, querySQL, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var projectID string
		var last sql.NullTime
		if err := rows.Scan(&projectID, &last); err != nil {
			return nil, err
		}
		if projectID != "" && last.Valid {
			out[projectID] = last.Time
		}
	}
	return out, rows.Err()
}

func (s *Service) validProjectIDs(ctx context.Context, querySQL, owner string) (map[string]struct{}, error) {
	rows, err := s.DB.QueryContext(ctx, querySQL, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]struct{}{}
	for rows.Next() {
		var projectID string
		if err := rows.Scan(&projectID); err != nil {
			return nil, err
		}
		out[projectID] = struct{}{}
	}
	return out, rows.Err()
}

func (s *Service) staleProjectIDs(ctx context.Context, owner string, module Module, valid map[string]struct{}) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT project_id FROM project_retention WHERE owner_username=? AND module=?`, owner, string(module))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stale := []string{}
	for rows.Next() {
		var projectID string
		if err := rows.Scan(&projectID); err != nil {
			return nil, err
		}
		if _, ok := valid[projectID]; !ok {
			stale = append(stale, projectID)
		}
	}
	return stale, rows.Err()
}

// archiveDue 把已到期的作品集移入归档；只更新状态，不删除任何数据。
func (s *Service) archiveDue(ctx context.Context, owner string) error {
	now := s.now().UTC()
	if _, err := s.DB.ExecContext(ctx, `UPDATE project_retention SET archived_at=?, updated_at=? WHERE owner_username=? AND module<>'novel_creation' AND archived_at IS NULL AND retention_expires_at<=?`, now, now, owner, now); err != nil {
		return err
	}
	// 调大保留天数后，尚未到期的归档作品会自动回到作品列表。
	_, err := s.DB.ExecContext(ctx, `UPDATE project_retention SET archived_at=NULL, updated_at=? WHERE owner_username=? AND module<>'novel_creation' AND manual_archived=0 AND archived_at IS NOT NULL AND retention_expires_at>?`, now, owner, now)
	if err != nil {
		return err
	}
	return s.archiveDueNovel(ctx, owner)
}

// ArchiveDue 在重算后统一执行到期归档。RunOnce 已包含该步骤，拆分出来便于测试。
func (s *Service) ArchiveDue(ctx context.Context, owner string) error {
	return s.archiveDue(ctx, owner)
}

// Status 返回该账号全部作品集的保留状态（含已归档）。
func (s *Service) Status(ctx context.Context, owner string) ([]Record, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("project retention database is not configured")
	}
	now := s.now().UTC()
	out := []Record{}
	rows, err := s.DB.QueryContext(ctx, `SELECT module, project_id, title, last_activity_at, retention_expires_at, archived_at FROM project_retention WHERE owner_username=?`, owner)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var record Record
		var module string
		var archived sql.NullTime
		if err := rows.Scan(&module, &record.ProjectID, &record.Title, &record.LastActivity, &record.ExpiresAt, &archived); err != nil {
			rows.Close()
			return nil, err
		}
		record.Module = Module(module)
		if archived.Valid {
			value := archived.Time
			record.ArchivedAt = &value
		}
		record.RemainingDays = remainingDays(record.ExpiresAt, now)
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range out {
		if out[i].Module == ModuleNovelCreation && out[i].ArchivedAt != nil {
			var status string
			e := s.DB.QueryRowContext(ctx, `SELECT status FROM novel_projects WHERE owner_username=? AND project_id=?`, owner, out[i].ProjectID).Scan(&status)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return nil, e
			}
			out[i].Purging = status == "purging"
		}
	}
	if s.Batch != nil {
		batches, err := s.Batch.ListBatchRetention(ctx, owner)
		if err != nil {
			return nil, err
		}
		for _, batch := range batches {
			last := batch.CompletedAt
			expires := batch.ExpiresAt.Time
			if !batch.ExpiresAt.Valid {
				// 首次 00:00 巡检前还没有写入到期时间，按设置实时计算展示值。
				expires = s.expiresAt(last, s.days(owner))
			}
			record := Record{
				Module:        ModuleBatchFactory,
				ProjectID:     batch.ID,
				Title:         batch.Title,
				LastActivity:  last,
				ExpiresAt:     expires,
				RemainingDays: remainingDays(expires, now),
			}
			if batch.HiddenAt.Valid {
				value := batch.HiddenAt.Time
				record.ArchivedAt = &value
			}
			out = append(out, record)
		}
	}
	return out, nil
}

// Archived 只返回已归档的作品集。
func (s *Service) Archived(ctx context.Context, owner string) ([]Record, error) {
	all, err := s.Status(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := []Record{}
	for _, record := range all {
		if record.ArchivedAt != nil {
			out = append(out, record)
		}
	}
	return out, nil
}

func validModule(module Module) bool {
	switch module {
	case ModuleBatchFactory, ModuleComicScreenplay, ModuleCMCanvas, ModuleNovelCreation:
		return true
	default:
		return false
	}
}

// Restore 把归档作品恢复到列表，并按当前保留天数从恢复时刻重新计时。
func (s *Service) Restore(ctx context.Context, owner string, module Module, projectID string) error {
	if s.DB == nil {
		return fmt.Errorf("project retention database is not configured")
	}
	if !validModule(module) || strings.TrimSpace(projectID) == "" {
		return ErrInvalid
	}
	if module == ModuleNovelCreation {
		return s.restoreNovel(ctx, owner, projectID)
	}
	now := s.now().UTC()
	days := s.days(owner)
	if module == ModuleBatchFactory {
		if s.Batch == nil {
			return ErrNotFound
		}
		return s.Batch.RestoreArchivedBatch(ctx, owner, projectID, days, now)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE project_retention SET last_activity_at=?, retention_expires_at=?, archived_at=NULL, updated_at=? WHERE owner_username=? AND module=? AND project_id=? AND archived_at IS NOT NULL`,
		now, now.AddDate(0, 0, days), now, owner, string(module), projectID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrNotFound
	}
	// 恢复后必须把作品本身的最后活动时间推到当前时刻，否则下一次 00:00 巡检
	// 会按旧的 updated_at 重新算出过期时间，导致刚恢复的作品又被归档。
	parentSQL := ""
	switch module {
	case ModuleComicScreenplay:
		parentSQL = `UPDATE screenplay_projects SET updated_at=? WHERE owner_username=? AND id=? AND deleted_at IS NULL`
	case ModuleCMCanvas:
		parentSQL = `UPDATE agent_projects SET updated_at=? WHERE owner_username=? AND id=? AND deleted_at IS NULL`
	case ModuleNovelCreation:
		parentSQL = `UPDATE novel_projects SET updated_at=? WHERE owner_username=? AND project_id=? AND deleted_at IS NULL`
	default:
		return ErrInvalid
	}
	touched, err := tx.ExecContext(ctx, parentSQL, now, owner, projectID)
	if err != nil {
		return err
	}
	if affected, _ := touched.RowsAffected(); affected != 1 {
		return ErrNotFound
	}
	return tx.Commit()
}

// Purge 彻底删除归档作品：批量工厂同时清理 TOS 与本地文件；其余模块删除数据库记录。
func (s *Service) Purge(ctx context.Context, owner string, module Module, projectID string) error {
	if s.DB == nil {
		return fmt.Errorf("project retention database is not configured")
	}
	if !validModule(module) || strings.TrimSpace(projectID) == "" {
		return ErrInvalid
	}
	if module == ModuleBatchFactory {
		if s.Batch == nil {
			return ErrNotFound
		}
		return s.Batch.PurgeArchivedBatch(ctx, owner, projectID, s.Local, s.Remote)
	}
	var archived sql.NullTime
	err := s.DB.QueryRowContext(ctx, `SELECT archived_at FROM project_retention WHERE owner_username=? AND module=? AND project_id=?`, owner, string(module), projectID).Scan(&archived)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !archived.Valid {
		return ErrInvalid
	}
	if module == ModuleNovelCreation {
		if err := s.beginNovelPurge(ctx, owner, projectID); err != nil {
			return err
		}
		if s.NovelObjects != nil {
			if err := s.NovelObjects(ctx, owner, projectID); err != nil {
				return err
			}
		}
	}
	if err := s.purgeModuleRows(ctx, owner, module, projectID); err != nil {
		return err
	}
	if module == ModuleNovelCreation {
		return nil
	}
	_, err = s.DB.ExecContext(ctx, `DELETE FROM project_retention WHERE owner_username=? AND module=? AND project_id=?`, owner, string(module), projectID)
	return err
}

func (s *Service) purgeModuleRows(ctx context.Context, owner string, module Module, projectID string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	switch module {
	case ModuleComicScreenplay:
		// 先删按批次关联的日志/步骤，再删项目级表，最后删项目本身。
		if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_batch_steps WHERE owner_username=? AND batch_id IN (SELECT id FROM screenplay_batches WHERE owner_username=? AND project_id=?)`, owner, owner, projectID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_batch_episodes WHERE owner_username=? AND batch_id IN (SELECT id FROM screenplay_batches WHERE owner_username=? AND project_id=?)`, owner, owner, projectID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_batch_logs WHERE owner_username=? AND batch_id IN (SELECT id FROM screenplay_batches WHERE owner_username=? AND project_id=?)`, owner, owner, projectID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_rewrite_manifests WHERE owner_username=? AND batch_id IN (SELECT id FROM screenplay_batches WHERE owner_username=? AND project_id=?)`, owner, owner, projectID); err != nil {
			return err
		}
		for _, table := range []string{
			"screenplay_asset_check_jobs", "screenplay_authoring", "screenplay_project_creative_configs",
			"screenplay_media_workbench", "screenplay_production_shots", "screenplay_production_jobs",
			"screenplay_documents", "screenplay_shots", "screenplay_characters", "screenplay_scenes",
			"screenplay_props", "screenplay_voices", "screenplay_asset_batches", "screenplay_assets",
			"screenplay_storyboards", "screenplay_storyboard_shots", "screenplay_batches",
			"screenplay_reviews", "screenplay_storyboard_generations",
		} {
			if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE owner_username=? AND project_id=?`, owner, projectID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_projects WHERE owner_username=? AND id=?`, owner, projectID); err != nil {
			return err
		}
	case ModuleCMCanvas:
		// 先删除画布素材的云端/本地对象，再清数据库记录。
		if s.CanvasAssets != nil && s.CanvasObjects != nil {
			assets, err := s.CanvasAssets.ListAssets(ctx, owner, projectID, "")
			if err != nil {
				return err
			}
			for _, asset := range assets {
				if err := s.CanvasObjects.Delete(ctx, asset); err != nil {
					return err
				}
			}
		}
		for _, table := range []string{
			"agent_generation_results", "agent_generation_tasks", "agent_video_jobs", "agent_shot_generation_links",
			"agent_script_shots", "agent_batch_factory_node_origins", "agent_batch_factory_bindings",
			"agent_executions", "agent_node_versions", "agent_canvas_edges", "agent_canvas_nodes",
			"agent_canvases", "agent_messages", "agent_assets",
		} {
			if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE owner_username=? AND project_id=?`, owner, projectID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM agent_projects WHERE owner_username=? AND id=?`, owner, projectID); err != nil {
			return err
		}
	case ModuleNovelCreation:
		var one int
		if err = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_projects WHERE owner_username=? AND project_id=? AND status='purging' FOR UPDATE`, owner, projectID).Scan(&one); err != nil {
			return err
		}
		var archived sql.NullTime
		if err = tx.QueryRowContext(ctx, `SELECT archived_at FROM project_retention WHERE owner_username=? AND module='novel_creation' AND project_id=? FOR UPDATE`, owner, projectID).Scan(&archived); err != nil {
			return err
		}
		if !archived.Valid {
			return ErrInvalid
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM novel_automation_rounds WHERE owner_username=? AND run_id IN (SELECT run_id FROM novel_runs WHERE owner_username=? AND project_id=?)`, owner, owner, projectID); err != nil {
			return err
		}
		// 小说表带 RESTRICT 外键，必须从子表往父表删。
		for _, table := range []string{
			"novel_content_records", "novel_run_dispatch", "novel_project_slots",
			"novel_run_events", "novel_run_steps", "novel_runs", "novel_schedules", "novel_batches",
			"novel_messages", "novel_chapter_versions", "novel_audits", "novel_revisions",
			"novel_chapters", "novel_volumes", "novel_memories", "novel_foundations",
			"novel_project_settings", "novel_materials", "novel_project_notes", "novel_sessions",
		} {
			if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE owner_username=? AND project_id=?`, owner, projectID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM novel_projects WHERE owner_username=? AND project_id=?`, owner, projectID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM project_retention WHERE owner_username=? AND module='novel_creation' AND project_id=?`, owner, projectID); err != nil {
			return err
		}
	default:
		return ErrInvalid
	}
	return tx.Commit()
}
