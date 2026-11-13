package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	// automationJobPageSize 限制每次只取一页 job id，避免一次性把整表 id 读进内存。
	automationJobPageSize = 500
	// automationJobPayloadWarnBytes 只做超限告警，不截断数据，避免丢失任务状态。
	automationJobPayloadWarnBytes = 1 << 20
)

// AutomationStateStore is the durable store for batch automation jobs.
// The coordinator keeps one in-memory snapshot and writes changes here; it
// must not re-read every job file on the dispatcher's one-second tick.
type AutomationStateStore interface {
	Load(context.Context) (automationStateFile, bool, error)
	Save(context.Context, automationStateFile) error
}

// AutomationJobSaver is implemented by stores that can upsert individual jobs.
// The coordinator writes one book progress event at a time; rewriting the
// whole jobs table for each event multiplies MySQL binlog volume by the job
// count and was the main reason binlogs grew by ~100MB every two minutes.
type AutomationJobSaver interface {
	SaveJobs(context.Context, map[string]map[string]any) error
}

type mysqlAutomationStateStore struct {
	db *sql.DB
}

func NewMySQLAutomationStateStore(db *sql.DB) AutomationStateStore {
	if db == nil {
		return nil
	}
	return &mysqlAutomationStateStore{db: db}
}

func (s *mysqlAutomationStateStore) Load(ctx context.Context) (automationStateFile, bool, error) {
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{}}
	if s == nil || s.db == nil {
		return state, false, nil
	}
	// 两步查询：先只对 id/updated_at 排序分页，再按 id 批量取 job_json。
	// 直接对含 job_json 的行 ORDER BY 会触发 MySQL filesort 1038（大 JSON 撑爆 sort buffer）。
	for offset := 0; ; offset += automationJobPageSize {
		ids, err := s.loadJobIDs(ctx, offset, automationJobPageSize)
		if err != nil {
			return state, false, err
		}
		if len(ids) == 0 {
			break
		}
		if err := s.loadJobsByID(ctx, state.Jobs, ids); err != nil {
			return state, false, err
		}
		if len(ids) < automationJobPageSize {
			break
		}
	}
	return state, len(state.Jobs) > 0, nil
}

func (s *mysqlAutomationStateStore) loadJobIDs(ctx context.Context, offset, limit int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM batch_automation_jobs ORDER BY updated_at, id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("load automation job ids: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan automation job id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate automation job ids: %w", err)
	}
	return ids, nil
}

func (s *mysqlAutomationStateStore) loadJobsByID(ctx context.Context, jobs map[string]map[string]any, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, job_json FROM batch_automation_jobs WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return fmt.Errorf("load automation jobs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return fmt.Errorf("scan automation job: %w", err)
		}
		if len(raw) > automationJobPayloadWarnBytes {
			log.Printf("[automation] job %s payload is %d bytes, exceeds %d; loaded as-is", id, len(raw), automationJobPayloadWarnBytes)
		}
		var job map[string]any
		if err := json.Unmarshal(raw, &job); err != nil {
			return fmt.Errorf("decode automation job %s: %w", id, err)
		}
		if job == nil {
			job = map[string]any{}
		}
		if strings.TrimSpace(automationStringValue(job["id"], "")) == "" {
			job["id"] = id
		}
		jobs[id] = job
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate automation jobs: %w", err)
	}
	return nil
}

func (s *mysqlAutomationStateStore) Save(ctx context.Context, state automationStateFile) error {
	return s.SaveJobs(ctx, state.Jobs)
}

// SaveJobs upserts only the jobs that actually changed. Callers must pass the
// complete job document for each id; unrelated rows are left untouched, so one
// book heartbeat no longer rewrites every other job's JSON.
func (s *mysqlAutomationStateStore) SaveJobs(ctx context.Context, jobs map[string]map[string]any) error {
	if s == nil || s.db == nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin automation save: %w", err)
	}
	const query = `INSERT INTO batch_automation_jobs
 (id, owner_username, batch_id, state, scheduled_at, next_retry_at, job_json, created_at, updated_at)
 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
 ON DUPLICATE KEY UPDATE owner_username=VALUES(owner_username), batch_id=VALUES(batch_id), state=VALUES(state),
 scheduled_at=VALUES(scheduled_at), next_retry_at=VALUES(next_retry_at), job_json=VALUES(job_json),
 created_at=VALUES(created_at), updated_at=VALUES(updated_at)`
	now := time.Now().UTC()
	for id, job := range jobs {
		if strings.TrimSpace(id) == "" {
			continue
		}
		raw, marshalErr := json.Marshal(job)
		if marshalErr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("encode automation job %s: %w", id, marshalErr)
		}
		createdAt := automationStoredJobTime(job["createdAt"])
		if createdAt.IsZero() {
			createdAt = now
		}
		updatedAt := automationStoredJobTime(job["updatedAt"])
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		if _, err := tx.ExecContext(ctx, query, id,
			automationStringValue(job["owner"], ""), automationStringValue(job["batchId"], ""),
			automationStringValue(job["state"], ""), automationNullableTime(job["scheduledAt"]),
			automationNextRetryAt(job), string(raw), createdAt, updatedAt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("save automation job %s: %w", id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit automation save: %w", err)
	}
	return nil
}

func automationStoredJobTime(value any) time.Time {
	switch typed := value.(type) {
	case time.Time:
		return typed
	case *time.Time:
		if typed != nil {
			return *typed
		}
	}
	text := strings.TrimSpace(automationStringValue(value, ""))
	if text == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999"} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func automationNullableTime(value any) any {
	if parsed := automationStoredJobTime(value); !parsed.IsZero() {
		return parsed
	}
	return nil
}

func automationNextRetryAt(job map[string]any) any {
	books, _ := job["books"].(map[string]any)
	var next time.Time
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil || !automationBoolValue(entry["retryRequested"]) {
			continue
		}
		at := automationStoredJobTime(entry["retryAt"])
		if at.IsZero() || (!next.IsZero() && !at.Before(next)) {
			continue
		}
		next = at
	}
	if next.IsZero() {
		return nil
	}
	return next
}
