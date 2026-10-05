package video

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
