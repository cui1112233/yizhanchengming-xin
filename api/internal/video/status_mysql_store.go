package video

import (
	"context"
	"fmt"
	"strings"
)

const qualifiedProductionJobSelect = `SELECT
 video_production_jobs.id,
 video_production_jobs.batch_project_id,
 video_production_jobs.book_id,
 video_production_jobs.status,
 video_production_jobs.input_revision,
 video_production_jobs.final_prompt_stage_run_id,
 video_production_jobs.final_prompt_version,
 video_production_jobs.final_prompt_text,
 video_production_jobs.provider,
 video_production_jobs.model,
 video_production_jobs.idempotency_key,
 video_production_jobs.error_code,
 video_production_jobs.error_message,
 video_production_jobs.created_at,
 video_production_jobs.updated_at
 FROM video_production_jobs`

func (s *MySQLStore) ListLatestProductionJobsByProject(ctx context.Context, projectID int64) ([]ProductionJob, error) {
	rows, err := s.db.QueryContext(ctx, qualifiedProductionJobSelect+`
 JOIN (
   SELECT book_id, MAX(id) AS latest_id
   FROM video_production_jobs
   WHERE batch_project_id=?
   GROUP BY book_id
 ) latest ON latest.latest_id=video_production_jobs.id
 ORDER BY video_production_jobs.book_id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ProductionJob, 0)
	for rows.Next() {
		var job ProductionJob
		if err := rows.Scan(productionJobScanArgs(&job)...); err != nil {
			return nil, err
		}
		out = append(out, job)
	}
	return out, rows.Err()
}

func (s *MySQLStore) ListProductionTasksByJobIDs(ctx context.Context, jobIDs []int64) (map[int64][]ProductionTask, error) {
	out := make(map[int64][]ProductionTask, len(jobIDs))
	if len(jobIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(jobIDs))
	args := make([]any, len(jobIDs))
	for i, id := range jobIDs {
		if id <= 0 {
			return nil, fmt.Errorf("video: invalid production job id")
		}
		placeholders[i] = "?"
		args[i] = id
	}
	rows, err := s.db.QueryContext(ctx, productionTaskSelect+` WHERE production_job_id IN (`+strings.Join(placeholders, ",")+`) ORDER BY production_job_id ASC, attempt ASC, id ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var task ProductionTask
		if err := rows.Scan(productionTaskScanArgs(&task)...); err != nil {
			return nil, err
		}
		out[task.ProductionJobID] = append(out[task.ProductionJobID], task)
	}
	return out, rows.Err()
}
