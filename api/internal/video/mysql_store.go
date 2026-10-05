package video

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) UpsertProviderConfig(ctx context.Context, cfg ProviderConfig) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("video: mysql store unavailable")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO video_provider_configs
(provider_key, model, create_url, tasks_url, result_url, encrypted_secret, secret_nonce, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
create_url=VALUES(create_url), tasks_url=VALUES(tasks_url), result_url=VALUES(result_url),
encrypted_secret=VALUES(encrypted_secret), secret_nonce=VALUES(secret_nonce), enabled=VALUES(enabled)`,
		cfg.ProviderKey, cfg.Model, cfg.CreateURL, cfg.TasksURL, cfg.ResultURL, cfg.EncryptedSecret, cfg.SecretNonce, cfg.Enabled,
	)
	return err
}

func (s *MySQLStore) GetProviderConfig(ctx context.Context, provider, model string) (ProviderConfig, error) {
	var cfg ProviderConfig
	err := s.db.QueryRowContext(ctx, `
SELECT id, provider_key, model, create_url, tasks_url, result_url, encrypted_secret, secret_nonce, enabled, created_at, updated_at
FROM video_provider_configs WHERE provider_key=? AND model=? LIMIT 1`, provider, model).Scan(
		&cfg.ID, &cfg.ProviderKey, &cfg.Model, &cfg.CreateURL, &cfg.TasksURL, &cfg.ResultURL,
		&cfg.EncryptedSecret, &cfg.SecretNonce, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ProviderConfig{}, providerError(ErrorProviderUnconfigured, "provider configuration not found", err)
	}
	return cfg, err
}

func (s *MySQLStore) CreateOrGetProductionJob(ctx context.Context, job ProductionJob) (ProductionJob, bool, error) {
	result, err := s.db.ExecContext(ctx, `
INSERT INTO video_production_jobs
(batch_project_id, book_id, status, input_revision, final_prompt_stage_run_id, final_prompt_version, final_prompt_text, provider, model, idempotency_key, error_code, error_message)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.BatchProjectID, job.BookID, job.Status, job.InputRevision, job.FinalPromptStageRunID,
		job.FinalPromptVersion, job.FinalPromptText, job.Provider, job.Model, job.IdempotencyKey, job.ErrorCode, job.ErrorMessage,
	)
	if err == nil {
		job.ID, err = result.LastInsertId()
		if err != nil {
			return ProductionJob{}, false, err
		}
		loaded, err := s.GetProductionJob(ctx, job.ID)
		return loaded, true, err
	}
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return ProductionJob{}, false, err
	}
	existing, err := s.getProductionJobByIdempotency(ctx, job.IdempotencyKey)
	return existing, false, err
}

func (s *MySQLStore) getProductionJobByIdempotency(ctx context.Context, key string) (ProductionJob, error) {
	var job ProductionJob
	err := s.db.QueryRowContext(ctx, productionJobSelect+` WHERE idempotency_key=? LIMIT 1`, key).Scan(productionJobScanArgs(&job)...)
	return job, err
}

const productionJobSelect = `SELECT id, batch_project_id, book_id, status, input_revision, final_prompt_stage_run_id, final_prompt_version, final_prompt_text, provider, model, idempotency_key, error_code, error_message, created_at, updated_at FROM video_production_jobs`

func productionJobScanArgs(job *ProductionJob) []any {
	return []any{&job.ID, &job.BatchProjectID, &job.BookID, &job.Status, &job.InputRevision, &job.FinalPromptStageRunID, &job.FinalPromptVersion, &job.FinalPromptText, &job.Provider, &job.Model, &job.IdempotencyKey, &job.ErrorCode, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt}
}

func (s *MySQLStore) GetProductionJob(ctx context.Context, id int64) (ProductionJob, error) {
	var job ProductionJob
	err := s.db.QueryRowContext(ctx, productionJobSelect+` WHERE id=? LIMIT 1`, id).Scan(productionJobScanArgs(&job)...)
	return job, err
}

func (s *MySQLStore) CreateProductionTask(ctx context.Context, task ProductionTask) (ProductionTask, error) {
	result, err := s.db.ExecContext(ctx, `
INSERT INTO video_production_tasks
(production_job_id, attempt, provider, model, request_id, provider_job_id, status, error_code, error_message, artifact_source_url, output_bucket, output_object_key, output_url)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ProductionJobID, task.Attempt, task.Provider, task.Model, task.RequestID, task.ProviderJobID,
		task.Status, task.ErrorCode, task.ErrorMessage, task.ArtifactSourceURL, task.OutputBucket, task.OutputObjectKey, task.OutputURL,
	)
	if err != nil {
		return ProductionTask{}, err
	}
	task.ID, err = result.LastInsertId()
	if err != nil {
		return ProductionTask{}, err
	}
	return s.GetProductionTask(ctx, task.ID)
}

const productionTaskSelect = `SELECT id, production_job_id, attempt, provider, model, request_id, provider_job_id, status, error_code, error_message, artifact_source_url, output_bucket, output_object_key, output_url, created_at, updated_at FROM video_production_tasks`

func productionTaskScanArgs(task *ProductionTask) []any {
	return []any{&task.ID, &task.ProductionJobID, &task.Attempt, &task.Provider, &task.Model, &task.RequestID, &task.ProviderJobID, &task.Status, &task.ErrorCode, &task.ErrorMessage, &task.ArtifactSourceURL, &task.OutputBucket, &task.OutputObjectKey, &task.OutputURL, &task.CreatedAt, &task.UpdatedAt}
}

func (s *MySQLStore) GetProductionTask(ctx context.Context, id int64) (ProductionTask, error) {
	var task ProductionTask
	err := s.db.QueryRowContext(ctx, productionTaskSelect+` WHERE id=? LIMIT 1`, id).Scan(productionTaskScanArgs(&task)...)
	return task, err
}

func (s *MySQLStore) LatestTaskForJob(ctx context.Context, jobID int64) (ProductionTask, error) {
	var task ProductionTask
	err := s.db.QueryRowContext(ctx, productionTaskSelect+` WHERE production_job_id=? ORDER BY attempt DESC, id DESC LIMIT 1`, jobID).Scan(productionTaskScanArgs(&task)...)
	return task, err
}

func (s *MySQLStore) UpdateProductionTask(ctx context.Context, task ProductionTask) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE video_production_tasks SET provider_job_id=?, status=?, error_code=?, error_message=?, artifact_source_url=?, output_bucket=?, output_object_key=?, output_url=? WHERE id=?`,
		task.ProviderJobID, task.Status, task.ErrorCode, task.ErrorMessage, task.ArtifactSourceURL,
		task.OutputBucket, task.OutputObjectKey, task.OutputURL, task.ID,
	)
	return err
}

func (s *MySQLStore) UpdateProductionJob(ctx context.Context, job ProductionJob) error {
	_, err := s.db.ExecContext(ctx, `UPDATE video_production_jobs SET status=?, error_code=?, error_message=? WHERE id=?`, job.Status, job.ErrorCode, job.ErrorMessage, job.ID)
	return err
}

func (s *MySQLStore) ListRecoverableTasks(ctx context.Context, limit int) ([]ProductionTask, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, productionTaskSelect+` WHERE status IN ('queued','running') AND provider_job_id<>'' ORDER BY id ASC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ProductionTask, 0)
	for rows.Next() {
		var task ProductionTask
		if err := rows.Scan(productionTaskScanArgs(&task)...); err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreateLocalExecutor(ctx context.Context, record LocalExecutorRecord) error {
	capabilities, err := json.Marshal(record.Capabilities)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO video_local_executors
(id, name, provider_key, model, capabilities_json, token_hash, last_seen_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, record.ID, record.Name, record.ProviderKey, record.Model, capabilities, record.TokenHash[:], record.LastSeenAt, record.CreatedAt, record.UpdatedAt)
	return err
}

const localExecutorSelect = `SELECT id, name, provider_key, model, capabilities_json, token_hash, last_seen_at, created_at, updated_at FROM video_local_executors`

func scanLocalExecutor(scan func(...any) error) (LocalExecutorRecord, error) {
	var record LocalExecutorRecord
	var capabilities []byte
	var tokenHash []byte
	if err := scan(&record.ID, &record.Name, &record.ProviderKey, &record.Model, &capabilities, &tokenHash, &record.LastSeenAt, &record.CreatedAt, &record.UpdatedAt); err != nil {
		return LocalExecutorRecord{}, err
	}
	if len(tokenHash) != len(record.TokenHash) {
		return LocalExecutorRecord{}, fmt.Errorf("video: invalid local executor token hash")
	}
	copy(record.TokenHash[:], tokenHash)
	if err := json.Unmarshal(capabilities, &record.Capabilities); err != nil {
		return LocalExecutorRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) GetLocalExecutorByTokenHash(ctx context.Context, hash [32]byte) (LocalExecutorRecord, error) {
	record, err := scanLocalExecutor(s.db.QueryRowContext(ctx, localExecutorSelect+` WHERE token_hash=? LIMIT 1`, hash[:]).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return LocalExecutorRecord{}, ErrLocalExecutorUnauthorized
	}
	return record, err
}

func (s *MySQLStore) UpdateLocalExecutorHeartbeat(ctx context.Context, id string, capabilities []string, now time.Time) error {
	payload, err := json.Marshal(capabilities)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE video_local_executors SET capabilities_json=?, last_seen_at=?, updated_at=? WHERE id=?`, payload, now, now, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrLocalExecutorUnauthorized
	}
	return nil
}

func (s *MySQLStore) ListLocalExecutors(ctx context.Context) ([]LocalExecutorRecord, error) {
	rows, err := s.db.QueryContext(ctx, localExecutorSelect+` ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LocalExecutorRecord, 0)
	for rows.Next() {
		record, err := scanLocalExecutor(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreateLocalExecutorTask(ctx context.Context, task LocalExecutorTask) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO video_local_executor_tasks
(id, source_task_id, provider_key, model, prompt, request_id, status, executor_id, artifact_url, error_code, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?, ?, ?)`, task.ID, task.SourceTaskID, task.ProviderKey, task.Model, task.Prompt, task.RequestID, task.Status, task.ArtifactURL, task.ErrorCode, task.ErrorMessage, task.CreatedAt, task.UpdatedAt)
	return err
}

const localExecutorTaskSelect = `SELECT id, source_task_id, provider_key, model, prompt, request_id, status, COALESCE(executor_id,''), artifact_url, error_code, error_message, created_at, updated_at FROM video_local_executor_tasks`

func scanLocalExecutorTask(scan func(...any) error) (LocalExecutorTask, error) {
	var task LocalExecutorTask
	if err := scan(&task.ID, &task.SourceTaskID, &task.ProviderKey, &task.Model, &task.Prompt, &task.RequestID, &task.Status, &task.ExecutorID, &task.ArtifactURL, &task.ErrorCode, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return LocalExecutorTask{}, err
	}
	return task, nil
}

func (s *MySQLStore) GetLocalExecutorTask(ctx context.Context, id string) (LocalExecutorTask, error) {
	task, err := scanLocalExecutorTask(s.db.QueryRowContext(ctx, localExecutorTaskSelect+` WHERE id=? LIMIT 1`, id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return LocalExecutorTask{}, ErrLocalExecutorTaskNotFound
	}
	return task, err
}

func (s *MySQLStore) CompleteLocalExecutorTask(ctx context.Context, id, executorID, artifactURL string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE video_local_executor_tasks SET status=?, executor_id=?, artifact_url=?, error_code='', error_message='', updated_at=? WHERE id=? AND status IN ('queued','running')`, TaskSucceeded, executorID, artifactURL, now, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	existing, err := s.GetLocalExecutorTask(ctx, id)
	if err != nil {
		return err
	}
	if existing.Status == TaskSucceeded && existing.ExecutorID == executorID && existing.ArtifactURL == artifactURL {
		return nil
	}
	return ErrLocalExecutorTaskNotFound
}

func (s *MySQLStore) FailLocalExecutorTask(ctx context.Context, id, executorID string, code ErrorCode, message string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE video_local_executor_tasks SET status=?, executor_id=?, error_code=?, error_message=?, updated_at=? WHERE id=? AND status IN ('queued','running')`, TaskFailed, executorID, code, message, now, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrLocalExecutorTaskNotFound
	}
	return nil
}

func (s *MySQLStore) CancelLocalExecutorTask(ctx context.Context, id string, now time.Time) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE video_local_executor_tasks SET status=?, updated_at=? WHERE id=? AND status IN ('queued','running')`, TaskCancelled, now, id)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
	}
	if _, err := s.GetLocalExecutorTask(ctx, id); err != nil {
		return false, err
	}
	return false, nil
}
