package generation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) PublishAdminPrompt(ctx context.Context, actor int64, key string, version int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE", key)
	if err != nil {
		return err
	}
	var targetID int64
	var targetContent string
	for rows.Next() {
		var id int64
		var currentVersion int
		var content string
		if err := rows.Scan(&id, &currentVersion, &content); err != nil {
			rows.Close()
			return err
		}
		if currentVersion == version {
			targetID = id
			targetContent = content
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE generation_prompts SET lifecycle='archived',enabled=0 WHERE prompt_key=? AND lifecycle='published'", key); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE generation_prompts SET lifecycle='published',enabled=1,published_at=UTC_TIMESTAMP(6),published_by_user_id=? WHERE prompt_key=? AND version=?", actor, key, version)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return fmt.Errorf("publish prompt version %d: %w", version, ErrNotFound)
	}
	h := sha256.Sum256([]byte(targetContent))
	requestID := observability.RequestID(ctx)
	if _, err = tx.ExecContext(ctx, "INSERT INTO admin_audit_logs(actor_user_id,capability,action,resource_type,resource_id,prompt_version_id,request_id,result,summary_json,content_sha256) VALUES(?,?,?,?,?,?,?,?,?,?)", actor, "admin.prompt.publish", "publish", "generation_prompt", key, targetID, requestID, "success", fmt.Sprintf("{\"version\":%d}", version), hex.EncodeToString(h[:])); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) RestoreAdminPrompt(ctx context.Context, actor int64, key string, source int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE", key)
	if err != nil {
		return err
	}
	defer rows.Close()
	max := 0
	var sourceID int64
	content := ""
	for rows.Next() {
		var id int64
		var v int
		var c string
		if err = rows.Scan(&id, &v, &c); err != nil {
			return err
		}
		if v > max {
			max = v
		}
		if v == source {
			sourceID = id
			content = c
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if content == "" || sourceID == 0 {
		return ErrNotFound
	}
	next := max + 1
	h := sha256.Sum256([]byte(content))
	contentSHA := hex.EncodeToString(h[:])
	result, err := tx.ExecContext(ctx, "INSERT INTO generation_prompts(prompt_key,version,content,enabled,lifecycle,seed_source,content_sha256) VALUES(?,?,?,0,'draft','restore',?)", key, next, content, contentSHA)
	if err != nil {
		return err
	}
	newID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE generation_prompts SET lifecycle='archived',enabled=0 WHERE prompt_key=? AND lifecycle='published'", key); err != nil {
		return err
	}
	result, err = tx.ExecContext(ctx, "UPDATE generation_prompts SET lifecycle='published',enabled=1,published_at=UTC_TIMESTAMP(6),published_by_user_id=? WHERE prompt_key=? AND version=?", actor, key, next)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return fmt.Errorf("restore prompt version %d: %w", next, ErrNotFound)
	}
	requestID := observability.RequestID(ctx)
	if _, err = tx.ExecContext(ctx, "INSERT INTO admin_audit_logs(actor_user_id,capability,action,resource_type,resource_id,prompt_version_id,request_id,result,summary_json,content_sha256) VALUES(?,?,?,?,?,?,?,?,?,?)", actor, "admin.prompt.publish", "restore", "generation_prompt", key, newID, requestID, "success", fmt.Sprintf("{\"source_version\":%d,\"new_version\":%d}", source, next), contentSHA); err != nil {
		return err
	}
	return tx.Commit()
}

func noRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MySQLStore) GetBookForProject(ctx context.Context, projectID, bookID int64) (intake.Book, error) {
	var b intake.Book
	err := s.db.QueryRowContext(ctx, `SELECT b.id,b.intake_id,b.source,b.platform_id,b.external_book_id,b.title,b.body_ref,b.original_text,b.category,b.genre,b.gender,b.gender_source,b.style,b.status,b.error_message,b.created_at,b.updated_at FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? AND b.id=?`, projectID, bookID).Scan(
		&b.ID, &b.IntakeID, &b.Source, &b.PlatformID, &b.ExternalBookID, &b.Title, &b.BodyRef, &b.OriginalText, &b.Category, &b.Genre, &b.Gender, &b.GenderSource, &b.Style, &b.Status, &b.ErrorMessage, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return intake.Book{}, noRows(err)
	}
	return b, nil
}

func (s *MySQLStore) ListBooksForProject(ctx context.Context, projectID int64) ([]intake.Book, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,b.intake_id,b.source,b.platform_id,b.external_book_id,b.title,b.body_ref,b.original_text,b.category,b.genre,b.gender,b.gender_source,b.style,b.status,b.error_message,b.created_at,b.updated_at FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? ORDER BY b.id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []intake.Book{}
	for rows.Next() {
		var b intake.Book
		if err := rows.Scan(&b.ID, &b.IntakeID, &b.Source, &b.PlatformID, &b.ExternalBookID, &b.Title, &b.BodyRef, &b.OriginalText, &b.Category, &b.Genre, &b.Gender, &b.GenderSource, &b.Style, &b.Status, &b.ErrorMessage, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *MySQLStore) ResolvePrompt(ctx context.Context, key string) (Prompt, error) {
	var p Prompt
	err := s.db.QueryRowContext(ctx, `SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts WHERE prompt_key=? AND enabled=1 AND lifecycle='published' ORDER BY version DESC LIMIT 1`, key).Scan(&p.ID, &p.Key, &p.Version, &p.Content, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Prompt{}, noRows(err)
	}
	return p, nil
}

func (s *MySQLStore) ListPrompts(ctx context.Context) ([]Prompt, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts ORDER BY prompt_key,version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Prompt{}
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.Key, &p.Version, &p.Content, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreateBookRun(ctx context.Context, v BookRun) (BookRun, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookRun{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE`, v.BatchProjectID).Scan(&archivedAt); err != nil {
		return BookRun{}, err
	}
	if archivedAt.Valid {
		return BookRun{}, ErrProjectArchived
	}
	r, err := tx.ExecContext(ctx, `INSERT INTO book_runs(batch_project_id,book_id,status,request_id,error_message,started_at,finished_at) VALUES(?,?,?,?,?,?,?)`, v.BatchProjectID, v.BookID, v.Status, v.RequestID, v.ErrorMessage, v.StartedAt, v.FinishedAt)
	if err != nil {
		return BookRun{}, fmt.Errorf("create book run: %w", err)
	}
	v.ID, _ = r.LastInsertId()
	if err := tx.Commit(); err != nil {
		return BookRun{}, err
	}
	return s.bookRunByID(ctx, v.ID)
}

func (s *MySQLStore) UpdateBookRun(ctx context.Context, v BookRun) (BookRun, error) {
	var runtimeRunID sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT run_id FROM book_runs WHERE id=?`, v.ID).Scan(&runtimeRunID); err != nil {
		return BookRun{}, noRows(err)
	}
	if runtimeRunID.Valid {
		return BookRun{}, ErrConflict
	}
	if v.Status != StatusRunning {
		r, err := s.db.ExecContext(ctx, `UPDATE book_runs SET status=?,request_id=?,error_message=?,started_at=?,finished_at=? WHERE id=?`, v.Status, v.RequestID, v.ErrorMessage, v.StartedAt, v.FinishedAt, v.ID)
		if err != nil {
			return BookRun{}, err
		}
		if n, _ := r.RowsAffected(); n == 0 {
			return BookRun{}, ErrNotFound
		}
		return s.bookRunByID(ctx, v.ID)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookRun{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE`, v.BatchProjectID).Scan(&archivedAt); err != nil {
		return BookRun{}, err
	}
	if archivedAt.Valid {
		return BookRun{}, ErrProjectArchived
	}
	r, err := tx.ExecContext(ctx, `UPDATE book_runs SET status=?,request_id=?,error_message=?,started_at=?,finished_at=? WHERE id=?`, v.Status, v.RequestID, v.ErrorMessage, v.StartedAt, v.FinishedAt, v.ID)
	if err != nil {
		return BookRun{}, err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return BookRun{}, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return BookRun{}, err
	}
	return s.bookRunByID(ctx, v.ID)
}

func (s *MySQLStore) bookRunByID(ctx context.Context, id int64) (BookRun, error) {
	var v BookRun
	var runID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT id,run_id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE id=?`, id).Scan(&v.ID, &runID, &v.BatchProjectID, &v.BookID, &v.Status, &v.RequestID, &v.ErrorMessage, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return BookRun{}, noRows(err)
	}
	if runID.Valid {
		v.RunID = runID.Int64
	}
	v.Status = NormalizeBookRunStatus(v.Status)
	return v, nil
}

func (s *MySQLStore) LatestBookRun(ctx context.Context, projectID, bookID int64) (BookRun, error) {
	var v BookRun
	var runID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT id,run_id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE batch_project_id=? AND book_id=? ORDER BY id DESC LIMIT 1`, projectID, bookID).Scan(&v.ID, &runID, &v.BatchProjectID, &v.BookID, &v.Status, &v.RequestID, &v.ErrorMessage, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return BookRun{}, noRows(err)
	}
	if runID.Valid {
		v.RunID = runID.Int64
	}
	v.Status = NormalizeBookRunStatus(v.Status)
	return v, nil
}

func (s *MySQLStore) ListBookRunsByProject(ctx context.Context, projectID int64) ([]BookRun, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,run_id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE batch_project_id=? ORDER BY id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BookRun{}
	for rows.Next() {
		var v BookRun
		var runID sql.NullInt64
		if err := rows.Scan(&v.ID, &runID, &v.BatchProjectID, &v.BookID, &v.Status, &v.RequestID, &v.ErrorMessage, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		if runID.Valid {
			v.RunID = runID.Int64
		}
		v.Status = NormalizeBookRunStatus(v.Status)
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreateStageRun(ctx context.Context, v StageRun) (StageRun, error) {
	r, err := s.db.ExecContext(ctx, `INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.BookRunID, v.BookID, v.Stage, v.Status, v.Attempt, v.RequestID, v.PromptKey, v.PromptVersion, v.InputSnapshot, v.OutputText, v.ErrorMessage, v.ValidationResult, v.StartedAt, v.FinishedAt)
	if err != nil {
		return StageRun{}, fmt.Errorf("create stage run: %w", err)
	}
	v.ID, _ = r.LastInsertId()
	return s.stageRunByID(ctx, v.ID)
}

func (s *MySQLStore) UpdateStageRun(ctx context.Context, v StageRun) (StageRun, error) {
	r, err := s.db.ExecContext(ctx, `UPDATE stage_runs SET status=?,request_id=?,prompt_key=?,prompt_version=?,input_snapshot=?,output_text=?,error_message=?,validation_result=?,started_at=?,finished_at=? WHERE id=?`, v.Status, v.RequestID, v.PromptKey, v.PromptVersion, v.InputSnapshot, v.OutputText, v.ErrorMessage, v.ValidationResult, v.StartedAt, v.FinishedAt, v.ID)
	if err != nil {
		return StageRun{}, err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return StageRun{}, ErrNotFound
	}
	return s.stageRunByID(ctx, v.ID)
}

func (s *MySQLStore) stageRunByID(ctx context.Context, id int64) (StageRun, error) {
	var v StageRun
	err := s.db.QueryRowContext(ctx, `SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE id=?`, id).Scan(&v.ID, &v.BookRunID, &v.BookID, &v.Stage, &v.Status, &v.Attempt, &v.RequestID, &v.PromptKey, &v.PromptVersion, &v.InputSnapshot, &v.OutputText, &v.ErrorMessage, &v.ValidationResult, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return StageRun{}, noRows(err)
	}
	return v, nil
}

func (s *MySQLStore) ListStageRuns(ctx context.Context, bookRunID int64) ([]StageRun, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE book_run_id=? ORDER BY id`, bookRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StageRun{}
	for rows.Next() {
		var v StageRun
		if err := rows.Scan(&v.ID, &v.BookRunID, &v.BookID, &v.Stage, &v.Status, &v.Attempt, &v.RequestID, &v.PromptKey, &v.PromptVersion, &v.InputSnapshot, &v.OutputText, &v.ErrorMessage, &v.ValidationResult, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *MySQLStore) LatestStageRun(ctx context.Context, bookRunID int64, stage Stage) (StageRun, error) {
	var v StageRun
	err := s.db.QueryRowContext(ctx, `SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE book_run_id=? AND stage=? ORDER BY attempt DESC,id DESC LIMIT 1`, bookRunID, stage).Scan(&v.ID, &v.BookRunID, &v.BookID, &v.Stage, &v.Status, &v.Attempt, &v.RequestID, &v.PromptKey, &v.PromptVersion, &v.InputSnapshot, &v.OutputText, &v.ErrorMessage, &v.ValidationResult, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return StageRun{}, noRows(err)
	}
	return v, nil
}

func (s *MySQLStore) AudioMeasurementByAsset(ctx context.Context, projectID, bookID int64, assetHash string) (AudioMeasurement, error) {
	var v AudioMeasurement
	err := s.db.QueryRowContext(ctx, `SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE batch_project_id=? AND book_id=? AND asset_hash=? LIMIT 1`, projectID, bookID, assetHash).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &v.AudioAsset, &v.AssetHash, &v.DurationMS, &v.MeasuredAt, &v.CreatedAt)
	if err != nil {
		return AudioMeasurement{}, noRows(err)
	}
	return v, nil
}

func (s *MySQLStore) LatestAudioMeasurement(ctx context.Context, projectID, bookID int64) (AudioMeasurement, error) {
	var v AudioMeasurement
	err := s.db.QueryRowContext(ctx, `SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE batch_project_id=? AND book_id=? ORDER BY measured_at DESC,id DESC LIMIT 1`, projectID, bookID).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &v.AudioAsset, &v.AssetHash, &v.DurationMS, &v.MeasuredAt, &v.CreatedAt)
	if err != nil {
		return AudioMeasurement{}, noRows(err)
	}
	return v, nil
}

func (s *MySQLStore) CreateAudioMeasurement(ctx context.Context, v AudioMeasurement) (AudioMeasurement, error) {
	r, err := s.db.ExecContext(ctx, `INSERT INTO audio_measurements(batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at) VALUES(?,?,?,?,?,?)`, v.BatchProjectID, v.BookID, v.AudioAsset, v.AssetHash, v.DurationMS, v.MeasuredAt)
	if err != nil {
		return AudioMeasurement{}, fmt.Errorf("create audio measurement: %w", err)
	}
	v.ID, _ = r.LastInsertId()
	err = s.db.QueryRowContext(ctx, `SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE id=?`, v.ID).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &v.AudioAsset, &v.AssetHash, &v.DurationMS, &v.MeasuredAt, &v.CreatedAt)
	if err != nil {
		return AudioMeasurement{}, err
	}
	return v, nil
}

func (s *MySQLStore) RuntimeBookRun(ctx context.Context, execution task9runtime.Execution) (BookRun, RuntimeExecutionInput, error) {
	var run BookRun
	var runID int64
	var attempt int
	var token uint64
	var owner, bookStatus, runStatus, runKind string
	var schemaVersion int
	var snapshot []byte
	var requestHash string
	var actorID int64
	var archivedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT br.id,br.run_id,br.batch_project_id,br.book_id,br.status,br.request_id,br.started_at,br.created_at,br.updated_at,br.attempt,br.execution_token,br.execution_owner,r.status,r.run_kind,r.request_schema_version,r.request_snapshot,r.request_hash,r.requested_by_user_id,p.archived_at FROM book_runs br JOIN runs r ON r.id=br.run_id JOIN batch_projects p ON p.id=r.batch_project_id WHERE br.id=?`, execution.BookRunID).Scan(
		&run.ID, &runID, &run.BatchProjectID, &run.BookID, &bookStatus, &run.RequestID, &run.StartedAt, &run.CreatedAt, &run.UpdatedAt, &attempt, &token, &owner, &runStatus, &runKind, &schemaVersion, &snapshot, &requestHash, &actorID, &archivedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BookRun{}, RuntimeExecutionInput{}, task9runtime.ErrStaleExecution
		}
		return BookRun{}, RuntimeExecutionInput{}, err
	}
	if archivedAt.Valid || runKind != "generation" || runStatus != string(task9runtime.RunRunning) || bookStatus != string(task9runtime.BookRunning) || attempt != execution.Attempt || token != execution.FencingToken || owner != execution.Owner {
		return BookRun{}, RuntimeExecutionInput{}, task9runtime.ErrStaleExecution
	}
	decoded, err := task9runtime.DecodeGenerationSnapshot(schemaVersion, snapshot, requestHash, run.BatchProjectID, actorID)
	if err != nil {
		return BookRun{}, RuntimeExecutionInput{}, ErrInvalid
	}
	selected := false
	for _, id := range decoded.BookIDs {
		if id == run.BookID {
			selected = true
			break
		}
	}
	if !selected {
		return BookRun{}, RuntimeExecutionInput{}, ErrInvalid
	}
	if decoded.Action == task9runtime.GenerationActionStageRetry {
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM book_runs source_br JOIN runs source_run ON source_run.id=source_br.run_id WHERE source_br.id=? AND source_br.batch_project_id=? AND source_br.book_id=? AND source_br.status='failed' AND source_run.batch_project_id=source_br.batch_project_id AND source_run.run_kind='generation' AND source_run.requested_by_user_id=? AND source_run.status IN ('failed','partial_failed') AND source_br.id<>?`, decoded.SourceBookRunID, run.BatchProjectID, run.BookID, decoded.RequestedByUserID, run.ID).Scan(&count); err != nil {
			return BookRun{}, RuntimeExecutionInput{}, err
		}
		if count != 1 {
			return BookRun{}, RuntimeExecutionInput{}, ErrInvalid
		}
	}
	run.RunID, run.Status = runID, StatusRunning
	projectConfig, _ := json.Marshal(struct {
		Constraints string `json:"constraints,omitempty"`
		Characters  string `json:"characters,omitempty"`
		Scenes      string `json:"scenes,omitempty"`
	}{Constraints: decoded.Config.Constraints, Characters: decoded.Config.Characters, Scenes: decoded.Config.Scenes})
	projectConfigText := string(projectConfig)
	if projectConfigText == "{}" {
		projectConfigText = ""
	}
	input := RuntimeExecutionInput{Action: decoded.Action, RetryStage: Stage(decoded.RetryStage), SourceBookRunID: decoded.SourceBookRunID, Request: RunBookRequest{BatchProjectID: run.BatchProjectID, BookID: run.BookID, HookEnabled: decoded.HookEnabled, PlotMode: decoded.PlotMode, DirectorMode: DirectorMode(decoded.DirectorMode), MatchAudio: decoded.MatchAudio, ShotDurationLimitSec: decoded.ShotDurationLimitSec, RequestID: run.RequestID, ProcessingRules: decoded.Config.ProcessingRulePromptRef, KnowledgeBase: decoded.Config.KnowledgePromptRef, ProjectConfig: projectConfigText, ModelConfig: decoded.Config.Model}}
	return run, input, nil
}

func (s *MySQLStore) runtimeParents(ctx context.Context, bookRunID int64) (int64, int64, error) {
	var runID sql.NullInt64
	var projectID int64
	if err := s.db.QueryRowContext(ctx, `SELECT run_id,batch_project_id FROM book_runs WHERE id=?`, bookRunID).Scan(&runID, &projectID); err != nil {
		return 0, 0, err
	}
	if !runID.Valid {
		return 0, 0, task9runtime.ErrStaleExecution
	}
	return runID.Int64, projectID, nil
}

func lockRuntimeExecution(ctx context.Context, tx *sql.Tx, runID, projectID int64, execution task9runtime.Execution) (int64, error) {
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE`, projectID).Scan(&archivedAt); err != nil {
		return 0, err
	}
	if archivedAt.Valid {
		return 0, task9runtime.ErrStaleExecution
	}
	var lockedProjectID int64
	var runStatus, runKind string
	if err := tx.QueryRowContext(ctx, `SELECT batch_project_id,status,run_kind FROM runs WHERE id=? FOR UPDATE`, runID).Scan(&lockedProjectID, &runStatus, &runKind); err != nil {
		return 0, err
	}
	if lockedProjectID != projectID || runStatus != string(task9runtime.RunRunning) || runKind != "generation" {
		return 0, task9runtime.ErrStaleExecution
	}
	var bookID int64
	var attempt int
	var token uint64
	var owner, status string
	if err := tx.QueryRowContext(ctx, `SELECT book_id,attempt,execution_token,execution_owner,status FROM book_runs WHERE id=? AND run_id=? AND batch_project_id=? FOR UPDATE`, execution.BookRunID, runID, projectID).Scan(&bookID, &attempt, &token, &owner, &status); err != nil {
		return 0, err
	}
	if attempt != execution.Attempt || token != execution.FencingToken || owner != execution.Owner || status != string(task9runtime.BookRunning) {
		return 0, task9runtime.ErrStaleExecution
	}
	return bookID, nil
}

func (s *MySQLStore) CreateStageRunFenced(ctx context.Context, execution task9runtime.Execution, v StageRun) (StageRun, error) {
	if !validRuntimeStage(v.Stage) {
		return StageRun{}, ErrInvalid
	}
	runID, projectID, err := s.runtimeParents(ctx, execution.BookRunID)
	if err != nil {
		return StageRun{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return StageRun{}, err
	}
	defer tx.Rollback()
	bookID, err := lockRuntimeExecution(ctx, tx, runID, projectID, execution)
	if err != nil {
		return StageRun{}, err
	}
	var maxAttempt sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(attempt) FROM stage_runs WHERE book_run_id=? AND stage=? FOR UPDATE`, execution.BookRunID, v.Stage).Scan(&maxAttempt); err != nil {
		return StageRun{}, err
	}
	v.BookRunID, v.BookID, v.Attempt = execution.BookRunID, bookID, 1
	if maxAttempt.Valid {
		v.Attempt = int(maxAttempt.Int64) + 1
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.BookRunID, v.BookID, v.Stage, v.Status, v.Attempt, v.RequestID, v.PromptKey, v.PromptVersion, v.InputSnapshot, v.OutputText, v.ErrorMessage, v.ValidationResult, v.StartedAt, v.FinishedAt)
	if err != nil {
		return StageRun{}, err
	}
	v.ID, err = result.LastInsertId()
	if err != nil {
		return StageRun{}, err
	}
	if err := scanStageRun(tx.QueryRowContext(ctx, `SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE id=?`, v.ID), &v); err != nil {
		return StageRun{}, err
	}
	if err := tx.Commit(); err != nil {
		return StageRun{}, err
	}
	return v, nil
}

func (s *MySQLStore) UpdateStageRunFenced(ctx context.Context, execution task9runtime.Execution, v StageRun) (StageRun, error) {
	runID, projectID, err := s.runtimeParents(ctx, execution.BookRunID)
	if err != nil {
		return StageRun{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return StageRun{}, err
	}
	defer tx.Rollback()
	bookID, err := lockRuntimeExecution(ctx, tx, runID, projectID, execution)
	if err != nil {
		return StageRun{}, err
	}
	var lockedBookRunID, lockedBookID int64
	if err := tx.QueryRowContext(ctx, `SELECT book_run_id,book_id FROM stage_runs WHERE id=? FOR UPDATE`, v.ID).Scan(&lockedBookRunID, &lockedBookID); err != nil {
		return StageRun{}, err
	}
	if lockedBookRunID != execution.BookRunID || lockedBookID != bookID || v.BookRunID != execution.BookRunID || (v.BookID != 0 && v.BookID != bookID) {
		return StageRun{}, task9runtime.ErrStaleExecution
	}
	v.BookID = bookID
	result, err := tx.ExecContext(ctx, `UPDATE stage_runs SET status=?,request_id=?,prompt_key=?,prompt_version=?,input_snapshot=?,output_text=?,error_message=?,validation_result=?,started_at=?,finished_at=? WHERE id=? AND book_run_id=? AND book_id=?`, v.Status, v.RequestID, v.PromptKey, v.PromptVersion, v.InputSnapshot, v.OutputText, v.ErrorMessage, v.ValidationResult, v.StartedAt, v.FinishedAt, v.ID, execution.BookRunID, bookID)
	if err != nil {
		return StageRun{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		if err != nil {
			return StageRun{}, err
		}
		return StageRun{}, task9runtime.ErrStaleExecution
	}
	if err := scanStageRun(tx.QueryRowContext(ctx, `SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE id=?`, v.ID), &v); err != nil {
		return StageRun{}, err
	}
	if err := tx.Commit(); err != nil {
		return StageRun{}, err
	}
	return v, nil
}

func scanStageRun(row interface{ Scan(...any) error }, v *StageRun) error {
	return row.Scan(&v.ID, &v.BookRunID, &v.BookID, &v.Stage, &v.Status, &v.Attempt, &v.RequestID, &v.PromptKey, &v.PromptVersion, &v.InputSnapshot, &v.OutputText, &v.ErrorMessage, &v.ValidationResult, &v.StartedAt, &v.FinishedAt, &v.CreatedAt, &v.UpdatedAt)
}
