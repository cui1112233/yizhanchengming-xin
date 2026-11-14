package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) CreateProject(ctx context.Context, project Project) (Project, error) {
	if project.AspectRatio == "" {
		project.AspectRatio = "9:16"
	}
	if !ValidCreativeAspectRatio(project.AspectRatio) {
		return Project{}, fmt.Errorf("%w: 不支持的剧本比例%s", ErrInvalidArgument, project.AspectRatio)
	}
	project.ID = strings.TrimSpace(project.ID)
	if project.ID == "" || strings.TrimSpace(project.OwnerUsername) == "" {
		return Project{}, fmt.Errorf("%w: 项目 ID 与 owner 不能为空", ErrInvalidArgument)
	}
	if project.PermissionMode == "" {
		project.PermissionMode = PermissionAuto
	}
	now := time.Now().UTC()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = now
	}
	project.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, `INSERT INTO screenplay_projects
(id, owner_username, title, script_mode, aspect_ratio, art_style, style_desc, selected_style, synopsis, source_text, model_id, prompt_id, asset_degree, permission_mode, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		project.ID, project.OwnerUsername, project.Title, project.ScriptMode, project.AspectRatio,
		project.ArtStyle, project.StyleDesc, project.SelectedStyle, project.Synopsis, project.SourceText,
		project.ModelID, project.PromptID, string(project.AssetDegree), string(project.PermissionMode), project.CreatedAt, project.UpdatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return Project{}, fmt.Errorf("%w: 项目 %s 已存在", ErrInvalidArgument, project.ID)
	}
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *MySQLStore) GetProject(ctx context.Context, owner, id string) (Project, error) {
	var project Project
	err := s.db.QueryRowContext(ctx, `SELECT id, owner_username, title, script_mode, aspect_ratio, art_style, style_desc, selected_style, synopsis, source_text, model_id, prompt_id, asset_degree, permission_mode, revision, created_at, updated_at
FROM screenplay_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).Scan(
		&project.ID, &project.OwnerUsername, &project.Title, &project.ScriptMode, &project.AspectRatio,
		&project.ArtStyle, &project.StyleDesc, &project.SelectedStyle, &project.Synopsis, &project.SourceText,
		&project.ModelID, &project.PromptID, &project.AssetDegree, &project.PermissionMode, &project.Revision, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	if err == nil {
		err = s.loadCreativeConfig(ctx, &project)
	}
	return project, err
}

func (s *MySQLStore) ListProjects(ctx context.Context, owner string) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, owner_username, title, script_mode, aspect_ratio, art_style, style_desc, selected_style, synopsis, model_id, prompt_id, asset_degree, permission_mode, revision, created_at, updated_at
FROM screenplay_projects WHERE owner_username = ? AND deleted_at IS NULL ORDER BY updated_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.OwnerUsername, &project.Title, &project.ScriptMode, &project.AspectRatio, &project.ArtStyle, &project.StyleDesc, &project.SelectedStyle, &project.Synopsis, &project.ModelID, &project.PromptID, &project.AssetDegree, &project.PermissionMode, &project.Revision, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (s *MySQLStore) UpdateProject(ctx context.Context, owner, id string, changes Project, expectedRevision int64) (Project, error) {
	if changes.AspectRatio != "" && !ValidCreativeAspectRatio(changes.AspectRatio) {
		return Project{}, fmt.Errorf("%w: 不支持的剧本比例%s", ErrInvalidArgument, changes.AspectRatio)
	}
	current, err := s.GetProject(ctx, owner, id)
	if err != nil {
		return Project{}, err
	}
	if current.Revision != expectedRevision {
		return Project{}, ErrRevisionConflict
	}
	if title := strings.TrimSpace(changes.Title); title != "" {
		current.Title = title
	}
	if changes.ScriptMode != "" {
		current.ScriptMode = changes.ScriptMode
	}
	if changes.AspectRatio != "" {
		current.AspectRatio = changes.AspectRatio
	}
	if changes.ArtStyle != "" {
		current.ArtStyle = changes.ArtStyle
	}
	current.StyleDesc, current.SelectedStyle, current.Synopsis = changes.StyleDesc, changes.SelectedStyle, changes.Synopsis
	if changes.ModelID != "" {
		current.ModelID = changes.ModelID
	}
	if changes.PromptID != "" {
		current.PromptID = changes.PromptID
	}
	if changes.AssetDegree != "" {
		current.AssetDegree = changes.AssetDegree
	}
	if changes.PermissionMode != "" {
		current.PermissionMode = changes.PermissionMode
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE screenplay_projects SET title=?, script_mode=?, aspect_ratio=?, art_style=?, style_desc=?, selected_style=?, synopsis=?, model_id=?, prompt_id=?, asset_degree=?, permission_mode=?, revision=revision+1, updated_at=? WHERE id=? AND owner_username=? AND deleted_at IS NULL AND revision=?`, current.Title, current.ScriptMode, current.AspectRatio, current.ArtStyle, current.StyleDesc, current.SelectedStyle, current.Synopsis, current.ModelID, current.PromptID, string(current.AssetDegree), string(current.PermissionMode), now, id, owner, expectedRevision)
	if err != nil {
		return Project{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Project{}, ErrRevisionConflict
	}
	return s.GetProject(ctx, owner, id)
}

func (s *MySQLStore) SaveProjectSource(ctx context.Context, owner, id, content string) (Project, error) {
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `UPDATE screenplay_projects SET source_text=?, updated_at=? WHERE id=? AND owner_username=? AND deleted_at IS NULL`, nullString(content), now, id, owner); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, owner, id)
}

func (s *MySQLStore) DeleteProject(ctx context.Context, owner, id string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE screenplay_projects SET deleted_at=?, revision=revision+1, updated_at=? WHERE id=? AND owner_username=? AND deleted_at IS NULL`, now, now, id, owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	return nil
}

func (s *MySQLStore) CreateBatch(ctx context.Context, batch *Batch) error {
	if batch == nil || batch.ID == "" || batch.ProjectID == "" || batch.OwnerUsername == "" {
		return fmt.Errorf("%w: 批次 ID / 项目 / owner 不能为空", ErrInvalidArgument)
	}
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = time.Now().UTC()
	}
	batch.UpdatedAt = batch.CreatedAt
	if batch.Steps == nil {
		batch.Steps = map[string]*StepState{}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pending, err := marshalJSON(batch.PendingActions)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_batches
(id, project_id, owner_username, session_id, status, current_step, current_episode, plan_version, permission_mode, pending_actions, last_error, revision, created_at, updated_at, last_run_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		batch.ID, batch.ProjectID, batch.OwnerUsername, batch.SessionID, string(batch.Status),
		batch.CurrentStep, batch.CurrentEpisode, batch.PlanVersion, string(batch.PermissionMode),
		jsonDocumentText(pending), batch.LastError, batch.Revision, batch.CreatedAt, batch.UpdatedAt, batch.LastRunAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return fmt.Errorf("%w: 项目 %s 已存在批次", ErrInvalidArgument, batch.ProjectID)
	}
	if err != nil {
		return err
	}
	if err := saveBatchSteps(ctx, tx, batch); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) GetBatch(ctx context.Context, owner, projectID string) (*Batch, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, project_id, owner_username, session_id, status, current_step, current_episode, plan_version, permission_mode, pending_actions, last_error, revision, created_at, updated_at, last_run_at
FROM screenplay_batches WHERE project_id = ? AND owner_username = ?`, projectID, owner)
	batch, err := scanBatch(row, owner)
	if err != nil {
		return nil, err
	}
	if err := loadBatchSteps(ctx, s.db, batch); err != nil {
		return nil, err
	}
	return batch, nil
}

// UpdateBatch 使用 revision 乐观锁写回。冲突时返回 ErrRevisionConflict，调用方应重新读取。
func (s *MySQLStore) UpdateBatch(ctx context.Context, batch *Batch, expectedRevision int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := updateBatchTx(ctx, tx, batch, expectedRevision); err != nil {
		return err
	}
	if err := saveBatchSteps(ctx, tx, batch); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	batch.Revision = expectedRevision + 1
	return nil
}

// MutateBatch 在事务内对批次行加显式行锁（SELECT ... FOR UPDATE），
// 保证同一批次的 run_next / cancel / 闸门更新不会并发丢写。
func (s *MySQLStore) MutateBatch(ctx context.Context, owner, projectID string, fn func(*Batch) error) (*Batch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `SELECT id, project_id, owner_username, session_id, status, current_step, current_episode, plan_version, permission_mode, pending_actions, last_error, revision, created_at, updated_at, last_run_at
FROM screenplay_batches WHERE project_id = ? AND owner_username = ? FOR UPDATE`, projectID, owner)
	batch, err := scanBatch(row, owner)
	if err != nil {
		return nil, err
	}
	if err := loadBatchSteps(ctx, tx, batch); err != nil {
		return nil, err
	}
	original := batch.Revision
	if err := fn(batch); err != nil {
		return nil, err
	}
	if err := updateBatchTx(ctx, tx, batch, original); err != nil {
		return nil, err
	}
	if err := saveBatchSteps(ctx, tx, batch); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	batch.Revision = original + 1
	return batch, nil
}

func updateBatchTx(ctx context.Context, tx *sql.Tx, batch *Batch, expectedRevision int64) error {
	pending, err := marshalJSON(batch.PendingActions)
	if err != nil {
		return err
	}
	batch.UpdatedAt = time.Now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE screenplay_batches SET
session_id = ?, status = ?, current_step = ?, current_episode = ?, plan_version = ?, permission_mode = ?,
pending_actions = ?, last_error = ?, revision = revision + 1, updated_at = ?, last_run_at = ?
WHERE id = ? AND owner_username = ? AND revision = ?`,
		batch.SessionID, string(NormalizeBatchStatus(batch.Status)), batch.CurrentStep, batch.CurrentEpisode,
		batch.PlanVersion, string(batch.PermissionMode), jsonDocumentText(pending), batch.LastError, batch.UpdatedAt,
		batch.LastRunAt, batch.ID, batch.OwnerUsername, expectedRevision)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrRevisionConflict
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBatch(row rowScanner, owner string) (*Batch, error) {
	var batch Batch
	var pending []byte
	var lastRun sql.NullTime
	err := row.Scan(&batch.ID, &batch.ProjectID, &batch.OwnerUsername, &batch.SessionID, &batch.Status,
		&batch.CurrentStep, &batch.CurrentEpisode, &batch.PlanVersion, &batch.PermissionMode,
		&pending, &batch.LastError, &batch.Revision, &batch.CreatedAt, &batch.UpdatedAt, &lastRun)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: 项目批次", ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	if lastRun.Valid {
		batch.LastRunAt = &lastRun.Time
	}
	if len(pending) > 0 {
		if err := json.Unmarshal(pending, &batch.PendingActions); err != nil {
			return nil, fmt.Errorf("解析 pending_actions 失败: %w", err)
		}
	}
	batch.OwnerUsername = owner
	batch.Steps = map[string]*StepState{}
	return &batch, nil
}

func loadBatchSteps(ctx context.Context, q queryer, batch *Batch) error {
	rows, err := q.QueryContext(ctx, `SELECT step_key, status, total_episodes, started_at, finished_at, artifacts, error, review
FROM screenplay_batch_steps WHERE batch_id = ? AND owner_username = ?`, batch.ID, batch.OwnerUsername)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		state := &StepState{}
		var started, finished sql.NullTime
		var artifacts, review []byte
		var stepErr sql.NullString
		if err := rows.Scan(&state.Key, &state.Status, &state.TotalEpisodes, &started, &finished, &artifacts, &stepErr, &review); err != nil {
			return err
		}
		if started.Valid {
			state.StartedAt = &started.Time
		}
		if finished.Valid {
			state.FinishedAt = &finished.Time
		}
		if stepErr.Valid {
			state.Error = stepErr.String
		}
		if len(artifacts) > 0 {
			if err := json.Unmarshal(artifacts, &state.Artifacts); err != nil {
				return fmt.Errorf("解析步骤 %s artifacts 失败: %w", state.Key, err)
			}
		}
		if len(review) > 0 {
			state.Review = &Review{}
			if err := json.Unmarshal(review, state.Review); err != nil {
				return fmt.Errorf("解析步骤 %s review 失败: %w", state.Key, err)
			}
		}
		batch.Steps[state.Key] = state
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// 补齐缺失的步键，避免旧数据导致状态机缺步。
	for _, def := range StepDefinitions {
		if _, ok := batch.Steps[def.Key]; !ok {
			batch.Steps[def.Key] = &StepState{Key: def.Key, Status: StepPending}
		}
	}
	return loadBatchEpisodes(ctx, q, batch)
}

func loadBatchEpisodes(ctx context.Context, q queryer, batch *Batch) error {
	rows, err := q.QueryContext(ctx, `SELECT step_key, episode_no, status, artifact, reason, updated_at
FROM screenplay_batch_episodes WHERE batch_id = ? AND owner_username = ? ORDER BY step_key, episode_no`, batch.ID, batch.OwnerUsername)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var stepKey string
		var ep EpisodeState
		var reason sql.NullString
		if err := rows.Scan(&stepKey, &ep.EpisodeNo, &ep.Status, &ep.Artifact, &reason, &ep.UpdatedAt); err != nil {
			return err
		}
		ep.Reason = reason.String
		step := batch.Steps[stepKey]
		if step == nil {
			step = &StepState{Key: stepKey, Status: StepPending}
			batch.Steps[stepKey] = step
		}
		if step.Episodes == nil {
			step.Episodes = map[int]*EpisodeState{}
		}
		copied := ep
		step.Episodes[ep.EpisodeNo] = &copied
	}
	return rows.Err()
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func saveBatchSteps(ctx context.Context, tx *sql.Tx, batch *Batch) error {
	for _, def := range StepDefinitions {
		state := batch.Steps[def.Key]
		if state == nil {
			continue
		}
		artifacts, err := marshalJSON(state.Artifacts)
		if err != nil {
			return err
		}
		var review []byte
		if state.Review != nil {
			review, err = marshalJSON(state.Review)
			if err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_batch_steps
(batch_id, owner_username, step_key, status, total_episodes, started_at, finished_at, artifacts, error, review)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE status = VALUES(status), total_episodes = VALUES(total_episodes),
started_at = VALUES(started_at), finished_at = VALUES(finished_at), artifacts = VALUES(artifacts),
error = VALUES(error), review = VALUES(review)`,
			batch.ID, batch.OwnerUsername, def.Key, string(state.Status), state.TotalEpisodes,
			state.StartedAt, state.FinishedAt, jsonDocumentText(artifacts), nullString(state.Error), nullableBytes(review))
		if err != nil {
			return err
		}
		for episodeNo, ep := range state.Episodes {
			if ep == nil {
				continue
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_batch_episodes
(batch_id, owner_username, step_key, episode_no, status, artifact, reason, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE status = VALUES(status), artifact = VALUES(artifact), reason = VALUES(reason), updated_at = VALUES(updated_at)`,
				batch.ID, batch.OwnerUsername, def.Key, episodeNo, string(ep.Status), ep.Artifact, nullString(ep.Reason), ep.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MySQLStore) AppendBatchLog(ctx context.Context, batchID, owner, level, message string) error {
	if strings.TrimSpace(level) == "" {
		level = "info"
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO screenplay_batch_logs (batch_id, owner_username, level, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		batchID, owner, level, message, time.Now().UTC())
	return err
}

func (s *MySQLStore) ListBatchLogs(ctx context.Context, batchID, owner string, limit int) ([]BatchLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, level, message, created_at FROM screenplay_batch_logs
WHERE batch_id = ? AND owner_username = ? ORDER BY id DESC LIMIT ?`, batchID, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]BatchLog, 0, limit)
	for rows.Next() {
		var entry BatchLog
		if err := rows.Scan(&entry.ID, &entry.Level, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}
	return logs, rows.Err()
}

// AcquireLock 通过唯一键 + 过期时间实现跨进程显式锁。
func (s *MySQLStore) AcquireLock(ctx context.Context, name, token string, ttl time.Duration) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(token) == "" {
		return fmt.Errorf("%w: 锁名与 token 不能为空", ErrInvalidArgument)
	}
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	now := time.Now().UTC()
	expires := now.Add(ttl)
	_, err := s.db.ExecContext(ctx, `INSERT INTO screenplay_locks (lock_name, owner_token, acquired_at, expires_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  owner_token = IF(expires_at < VALUES(acquired_at), VALUES(owner_token), owner_token),
  acquired_at = IF(expires_at < VALUES(acquired_at), VALUES(acquired_at), acquired_at),
  expires_at = IF(expires_at < VALUES(acquired_at), VALUES(expires_at), expires_at)`,
		name, token, now, expires)
	if err != nil {
		return err
	}
	var holder string
	if err := s.db.QueryRowContext(ctx, `SELECT owner_token FROM screenplay_locks WHERE lock_name = ?`, name).Scan(&holder); err != nil {
		return err
	}
	if holder != token {
		return fmt.Errorf("%w: %s", ErrLockBusy, name)
	}
	return nil
}

func (s *MySQLStore) ReleaseLock(ctx context.Context, name, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM screenplay_locks WHERE lock_name = ? AND owner_token = ?`, name, token)
	return err
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func jsonDocumentText(value []byte) string {
	if len(value) == 0 {
		return "null"
	}
	return string(value)
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
