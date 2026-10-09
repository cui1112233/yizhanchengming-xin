package task9runtime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrIdempotencyConflict = errors.New("generation idempotency conflict")
var ErrInvalidGenerationRequest = errors.New("invalid generation request")

type GenerationRequest struct {
	BatchProjectID       int64
	BookIDs              []int64
	HookEnabled          bool
	PlotMode             bool
	DirectorMode         string
	MatchAudio           bool
	ShotDurationLimitSec int64
	RequestID            string
	RequestedByUserID    int64
	Action               string
	RetryStage           string
	SourceBookRunID      int64
}

const (
	GenerationActionFull       = "full"
	GenerationActionStageRetry = "stage_retry"
)

type AdmissionResult struct {
	Run      RunRecord
	Items    []WorkItem
	TaskIDs  []int64
	Created  bool
	Dispatch string
}

var ErrExecutorUnavailable = errors.New("generation executor unavailable")

type GenerationTaskStatus struct {
	TaskID  int64  `json:"taskId"`
	BookID  int64  `json:"bookId"`
	Attempt int    `json:"attempt"`
	Status  string `json:"status"`
}

type GenerationRunStatus struct {
	RunID          int64                  `json:"runId"`
	BatchProjectID int64                  `json:"batchProjectId"`
	Status         string                 `json:"status"`
	Terminal       bool                   `json:"terminal"`
	Counts         GenerationRunCounts    `json:"counts"`
	Tasks          []GenerationTaskStatus `json:"tasks"`
}

type GenerationRunCounts struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type GenerationAdmissionStore interface {
	AdmitGeneration(context.Context, GenerationRequest) (AdmissionResult, error)
	GenerationRun(context.Context, int64, int64) (GenerationRunStatus, error)
}

// RuntimeReadiness is implemented by the single supervised generation runtime.
// Admission checks it on every request so a stopped or degraded worker cannot
// continue accepting work merely because it was ready when the app started.
type RuntimeReadiness interface {
	Ready() bool
}

type StaticRuntimeReadiness bool

func (r StaticRuntimeReadiness) Ready() bool { return bool(r) }

type GenerationAdmissionService struct {
	store       GenerationAdmissionStore
	coordinator RuntimeCoordinator
	readiness   RuntimeReadiness
}

func NewGenerationAdmissionService(store GenerationAdmissionStore, coordinator RuntimeCoordinator, readiness RuntimeReadiness) *GenerationAdmissionService {
	return &GenerationAdmissionService{store: store, coordinator: coordinator, readiness: readiness}
}

func (s *GenerationAdmissionService) Ready() bool {
	return s != nil && s.store != nil && s.coordinator != nil && s.readiness != nil && s.readiness.Ready()
}

func (s *GenerationAdmissionService) AdmitGeneration(ctx context.Context, req GenerationRequest) (AdmissionResult, error) {
	if !s.Ready() {
		return AdmissionResult{}, ErrExecutorUnavailable
	}
	result, err := s.store.AdmitGeneration(ctx, req)
	if err != nil {
		return AdmissionResult{}, err
	}
	result.Dispatch = "queued"
	for _, item := range result.Items {
		if err := s.coordinator.Enqueue(ctx, item); err != nil {
			// Admission is already durable. Recovery owns dispatch repair; the HTTP
			// contract must not turn a committed Run into an ambiguous 5xx.
			result.Dispatch = "pending"
			break
		}
	}
	return result, nil
}

func (s *GenerationAdmissionService) GenerationRun(ctx context.Context, projectID, runID int64) (GenerationRunStatus, error) {
	if s == nil || s.store == nil {
		return GenerationRunStatus{}, ErrExecutorUnavailable
	}
	return s.store.GenerationRun(ctx, projectID, runID)
}

// GenerationSnapshot is the immutable, versioned execution input. Provider
// credentials, source text, client durations and Prompt content cannot enter it.
type GenerationSnapshot struct {
	SchemaVersion        int                      `json:"schema_version"`
	BatchProjectID       int64                    `json:"batch_project_id"`
	BookIDs              []int64                  `json:"book_ids"`
	SelectionAll         bool                     `json:"selection_all"`
	HookEnabled          bool                     `json:"hook_enabled"`
	PlotMode             bool                     `json:"plot_mode"`
	DirectorMode         string                   `json:"director_mode"`
	MatchAudio           bool                     `json:"match_audio"`
	ShotDurationLimitSec int64                    `json:"shot_duration_limit_sec"`
	RequestedByUserID    int64                    `json:"requested_by_user_id"`
	Action               string                   `json:"action,omitempty"`
	RetryStage           string                   `json:"retry_stage,omitempty"`
	SourceBookRunID      int64                    `json:"source_book_run_id,omitempty"`
	Config               GenerationConfigSnapshot `json:"config"`
}

type GenerationConfigSnapshot struct {
	ProcessingRulePromptRef string `json:"processing_rule_prompt_ref,omitempty"`
	KnowledgePromptRef      string `json:"knowledge_prompt_ref,omitempty"`
	Constraints             string `json:"constraints,omitempty"`
	Characters              string `json:"characters,omitempty"`
	Scenes                  string `json:"scenes,omitempty"`
	Model                   string `json:"model,omitempty"`
}

const generationConfigFieldLimit = 16 * 1024
const generationConfigTotalLimit = 64 * 1024

type storedGenerationSettings struct {
	Production struct {
		ScriptWorkspace struct {
			Constraints string `json:"constraints"`
			Characters  string `json:"characters"`
			Scenes      string `json:"scenes"`
			Model       string `json:"model"`
		} `json:"scriptWorkspace"`
	} `json:"production"`
	ProcessingRulePromptRef string `json:"processingRulePromptRef"`
	KnowledgePromptRef      string `json:"knowledgePromptRef"`
}

func generationConfigFromSettings(projectRaw, profileRaw []byte, profileID int64, profileName, profileVersion string) (GenerationConfigSnapshot, error) {
	var project, profile storedGenerationSettings
	if len(projectRaw) > 0 && string(projectRaw) != "null" {
		if err := json.Unmarshal(projectRaw, &project); err != nil {
			return GenerationConfigSnapshot{}, ErrInvalidGenerationRequest
		}
	}
	if len(profileRaw) > 0 && string(profileRaw) != "null" {
		if err := json.Unmarshal(profileRaw, &profile); err != nil {
			return GenerationConfigSnapshot{}, ErrInvalidGenerationRequest
		}
	}
	choose := func(projectValue, profileValue string) string {
		if strings.TrimSpace(projectValue) != "" {
			return strings.TrimSpace(projectValue)
		}
		return strings.TrimSpace(profileValue)
	}
	workspace := project.Production.ScriptWorkspace
	profileWorkspace := profile.Production.ScriptWorkspace
	_ = profileID
	_ = profileName
	_ = profileVersion
	config := GenerationConfigSnapshot{
		ProcessingRulePromptRef: choose(project.ProcessingRulePromptRef, profile.ProcessingRulePromptRef),
		KnowledgePromptRef:      choose(project.KnowledgePromptRef, profile.KnowledgePromptRef),
		Constraints:             choose(workspace.Constraints, profileWorkspace.Constraints),
		Characters:              choose(workspace.Characters, profileWorkspace.Characters),
		Scenes:                  choose(workspace.Scenes, profileWorkspace.Scenes),
		Model:                   choose(workspace.Model, profileWorkspace.Model),
	}
	if err := validateGenerationConfig(config); err != nil {
		return GenerationConfigSnapshot{}, err
	}
	return config, nil
}

func validateGenerationConfig(config GenerationConfigSnapshot) error {
	total := 0
	for _, value := range []string{config.ProcessingRulePromptRef, config.KnowledgePromptRef, config.Constraints, config.Characters, config.Scenes, config.Model} {
		if !utf8.ValidString(value) || len(value) > generationConfigFieldLimit {
			return ErrInvalidGenerationRequest
		}
		total += len(value)
	}
	if total > generationConfigTotalLimit {
		return ErrInvalidGenerationRequest
	}
	return nil
}

func normalizeGenerationRequest(r GenerationRequest) (GenerationSnapshot, error) {
	if r.BatchProjectID <= 0 || r.RequestedByUserID <= 0 || strings.TrimSpace(r.RequestID) == "" || !utf8.ValidString(r.RequestID) || utf8.RuneCountInString(r.RequestID) > 191 {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	mode := r.DirectorMode
	if mode == "" {
		mode = "normal"
	}
	if mode != "normal" && mode != "h3" {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	duration := r.ShotDurationLimitSec
	if duration == 0 {
		duration = 15
	}
	if duration != 10 && duration != 15 {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	ids := append([]int64{}, r.BookIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	unique := ids[:0]
	for _, id := range ids {
		if id <= 0 {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	action := strings.TrimSpace(r.Action)
	if action == "" {
		action = GenerationActionFull
	}
	retryStage := strings.TrimSpace(r.RetryStage)
	switch action {
	case GenerationActionFull:
		if retryStage != "" || r.SourceBookRunID != 0 {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
	case GenerationActionStageRetry:
		if len(unique) != 1 || len(r.BookIDs) != 1 || r.SourceBookRunID <= 0 || !isRetryStage(retryStage) {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
	default:
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	return GenerationSnapshot{SchemaVersion: 2, BatchProjectID: r.BatchProjectID, BookIDs: unique, SelectionAll: len(r.BookIDs) == 0, HookEnabled: r.HookEnabled, PlotMode: r.PlotMode, DirectorMode: mode, MatchAudio: r.MatchAudio, ShotDurationLimitSec: duration, RequestedByUserID: r.RequestedByUserID, Action: action, RetryStage: retryStage, SourceBookRunID: r.SourceBookRunID}, nil
}

func isRetryStage(stage string) bool {
	switch stage {
	case "SCRIPT", "HOOK", "DIRECTOR", "FINAL_PROMPT":
		return true
	default:
		return false
	}
}

type generationSnapshotV1 struct {
	SchemaVersion        int     `json:"schema_version"`
	BatchProjectID       int64   `json:"batch_project_id"`
	BookIDs              []int64 `json:"book_ids"`
	SelectionAll         bool    `json:"selection_all"`
	HookEnabled          bool    `json:"hook_enabled"`
	PlotMode             bool    `json:"plot_mode"`
	DirectorMode         string  `json:"director_mode"`
	MatchAudio           bool    `json:"match_audio"`
	ShotDurationLimitSec int64   `json:"shot_duration_limit_sec"`
	RequestedByUserID    int64   `json:"requested_by_user_id"`
}

func encodeGenerationSnapshot(s GenerationSnapshot) ([]byte, string, error) {
	var value any = s
	if s.SchemaVersion == 1 {
		value = generationSnapshotV1{SchemaVersion: 1, BatchProjectID: s.BatchProjectID, BookIDs: s.BookIDs, SelectionAll: s.SelectionAll, HookEnabled: s.HookEnabled, PlotMode: s.PlotMode, DirectorMode: s.DirectorMode, MatchAudio: s.MatchAudio, ShotDurationLimitSec: s.ShotDurationLimitSec, RequestedByUserID: s.RequestedByUserID}
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(b)
	return b, hex.EncodeToString(sum[:]), nil
}

func decodeGenerationSnapshot(version int, b []byte, hash string, projectID, actorID int64) (GenerationSnapshot, error) {
	if version != 1 && version != 2 {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	var s GenerationSnapshot
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if version == 1 {
		var legacy generationSnapshotV1
		if err := d.Decode(&legacy); err != nil {
			return s, ErrInvalidGenerationRequest
		}
		s = GenerationSnapshot{SchemaVersion: legacy.SchemaVersion, BatchProjectID: legacy.BatchProjectID, BookIDs: legacy.BookIDs, SelectionAll: legacy.SelectionAll, HookEnabled: legacy.HookEnabled, PlotMode: legacy.PlotMode, DirectorMode: legacy.DirectorMode, MatchAudio: legacy.MatchAudio, ShotDurationLimitSec: legacy.ShotDurationLimitSec, RequestedByUserID: legacy.RequestedByUserID, Action: GenerationActionFull}
	} else if err := d.Decode(&s); err != nil {
		return s, ErrInvalidGenerationRequest
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return s, ErrInvalidGenerationRequest
	}
	if s.SchemaVersion != version || s.BatchProjectID != projectID || s.RequestedByUserID != actorID || len(s.BookIDs) == 0 {
		return s, ErrInvalidGenerationRequest
	}
	normalized, err := normalizeGenerationRequest(GenerationRequest{BatchProjectID: s.BatchProjectID, BookIDs: s.BookIDs, HookEnabled: s.HookEnabled, PlotMode: s.PlotMode, DirectorMode: s.DirectorMode, MatchAudio: s.MatchAudio, ShotDurationLimitSec: s.ShotDurationLimitSec, RequestID: "snapshot", RequestedByUserID: s.RequestedByUserID, Action: s.Action, RetryStage: s.RetryStage, SourceBookRunID: s.SourceBookRunID})
	normalized.SelectionAll = s.SelectionAll
	if version == 1 {
		normalized.SchemaVersion = 1
	} else {
		if err := validateGenerationConfig(s.Config); err != nil {
			return s, err
		}
		normalized.Config = s.Config
	}
	if err != nil || !reflect.DeepEqual(s, normalized) {
		return s, ErrInvalidGenerationRequest
	}
	_, canonicalHash, err := encodeGenerationSnapshot(s)
	if err != nil || canonicalHash != hash {
		return s, ErrInvalidGenerationRequest
	}
	return s, nil
}

// DecodeGenerationSnapshot validates and decodes the immutable admission
// snapshot for domain executors. It does not expose database internals or
// allow callers to bypass canonical hash validation.
func DecodeGenerationSnapshot(version int, b []byte, hash string, projectID, actorID int64) (GenerationSnapshot, error) {
	return decodeGenerationSnapshot(version, b, hash, projectID, actorID)
}

const generationRunColumns = `id,batch_project_id,idempotency_key,run_at,status,max_attempts,run_kind,target_book_id,request_schema_version,request_snapshot,request_hash,requested_by_user_id`

type rowScanner interface{ Scan(...any) error }

func scanGenerationRun(row rowScanner) (RunRecord, error) {
	var r RunRecord
	var status string
	var target, actor sql.NullInt64
	var key sql.NullString
	var snapshot []byte
	err := row.Scan(&r.ID, &r.BatchProjectID, &key, &r.RunAt, &status, &r.MaxAttempts, &r.RunKind, &target, &r.RequestSchemaVersion, &snapshot, &r.RequestHash, &actor)
	r.RequestSnapshot = json.RawMessage(snapshot)
	r.IdempotencyKey = key.String
	r.Status = RunState(status)
	r.RequestedByUserID = actor.Int64
	if target.Valid {
		r.TargetBookID = &target.Int64
	}
	return r, err
}

func lockGenerationProject(ctx context.Context, tx *sql.Tx, projectID int64, skipLocked bool) (int64, error) {
	query := `SELECT intake_id,archived_at FROM batch_projects WHERE id=? FOR UPDATE`
	if skipLocked {
		query += ` SKIP LOCKED`
	}
	var intakeID int64
	var archived sql.NullTime
	if err := tx.QueryRowContext(ctx, query, projectID).Scan(&intakeID, &archived); err != nil {
		return 0, err
	}
	if archived.Valid {
		return 0, ErrProjectArchived
	}
	return intakeID, nil
}

func projectBookIDs(ctx context.Context, tx *sql.Tx, intakeID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM books WHERE intake_id=? ORDER BY id`, intakeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
func validateFrozenBooks(ctx context.Context, tx *sql.Tx, intakeID int64, ids []int64) error {
	owned, err := projectBookIDs(ctx, tx, intakeID)
	if err != nil {
		return err
	}
	allowed := make(map[int64]bool, len(owned))
	for _, id := range owned {
		allowed[id] = true
	}
	if len(ids) == 0 {
		return ErrInvalidGenerationRequest
	}
	for _, id := range ids {
		if !allowed[id] {
			return ErrInvalidGenerationRequest
		}
	}
	return nil
}

func loadGenerationConfig(ctx context.Context, tx *sql.Tx, projectID int64) (GenerationConfigSnapshot, error) {
	var projectRaw, profileRaw []byte
	var profileID sql.NullInt64
	var profileName, profileVersion sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(ps.settings_json,JSON_OBJECT()),vp.id,vp.profile_name,vp.version,COALESCE(vp.settings_json,JSON_OBJECT()) FROM batch_projects p LEFT JOIN batch_project_settings ps ON ps.batch_project_id=p.id LEFT JOIN batch_version_config_profiles vp ON vp.batch_project_id=p.id WHERE p.id=?`, projectID).Scan(&projectRaw, &profileID, &profileName, &profileVersion, &profileRaw)
	if err != nil {
		return GenerationConfigSnapshot{}, err
	}
	return generationConfigFromSettings(projectRaw, profileRaw, profileID.Int64, profileName.String, profileVersion.String)
}

func materializeGenerationBooks(ctx context.Context, tx *sql.Tx, r RunRecord, snapshot GenerationSnapshot) ([]WorkItem, error) {
	for _, bookID := range snapshot.BookIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO book_runs (run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status,request_id) VALUES (?,?,?,1,?,1,'queued',?) ON DUPLICATE KEY UPDATE id=id`, r.ID, r.BatchProjectID, bookID, r.MaxAttempts, r.IdempotencyKey); err != nil {
			return nil, err
		}
	}
	return queuedGenerationItems(ctx, tx, r.ID)
}
func queuedGenerationItems(ctx context.Context, tx *sql.Tx, runID int64) ([]WorkItem, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,book_id,attempt FROM book_runs WHERE run_id=? AND attempt=1 AND status='queued' ORDER BY book_id FOR UPDATE`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WorkItem
	for rows.Next() {
		var w WorkItem
		if err := rows.Scan(&w.BookRunID, &w.BookID, &w.Attempt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func loadStageRetrySnapshot(ctx context.Context, tx *sql.Tx, requested GenerationSnapshot) (GenerationSnapshot, error) {
	if requested.Action != GenerationActionStageRetry || len(requested.BookIDs) != 1 {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	var sourceRunID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT run_id FROM book_runs WHERE id=?`, requested.SourceBookRunID).Scan(&sourceRunID); err != nil || !sourceRunID.Valid {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return GenerationSnapshot{}, err
		}
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	sourceRun, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, sourceRunID.Int64))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
		return GenerationSnapshot{}, err
	}
	if sourceRun.BatchProjectID != requested.BatchProjectID || sourceRun.RunKind != "generation" || sourceRun.RequestedByUserID != requested.RequestedByUserID || (sourceRun.Status != RunFailed && sourceRun.Status != RunPartialFailed) {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	sourceSnapshot, err := decodeGenerationSnapshot(sourceRun.RequestSchemaVersion, sourceRun.RequestSnapshot, sourceRun.RequestHash, sourceRun.BatchProjectID, sourceRun.RequestedByUserID)
	if err != nil {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	bookID := requested.BookIDs[0]
	frozenBook := false
	for _, id := range sourceSnapshot.BookIDs {
		if id == bookID {
			frozenBook = true
			break
		}
	}
	if !frozenBook {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	var sourceBookID int64
	var sourceBookStatus string
	if err := tx.QueryRowContext(ctx, `SELECT book_id,status FROM book_runs WHERE id=? AND run_id=? AND batch_project_id=? FOR UPDATE`, requested.SourceBookRunID, sourceRun.ID, requested.BatchProjectID).Scan(&sourceBookID, &sourceBookStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
		return GenerationSnapshot{}, err
	}
	if sourceBookID != bookID || sourceBookStatus != string(BookFailed) {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	var stageStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM stage_runs WHERE book_run_id=? AND stage=? ORDER BY attempt DESC,id DESC LIMIT 1 FOR UPDATE`, requested.SourceBookRunID, requested.RetryStage).Scan(&stageStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GenerationSnapshot{}, ErrInvalidGenerationRequest
		}
		return GenerationSnapshot{}, err
	}
	if stageStatus != "failed" {
		return GenerationSnapshot{}, ErrInvalidGenerationRequest
	}
	requested.HookEnabled = sourceSnapshot.HookEnabled
	requested.PlotMode = sourceSnapshot.PlotMode
	requested.DirectorMode = sourceSnapshot.DirectorMode
	requested.MatchAudio = sourceSnapshot.MatchAudio
	requested.ShotDurationLimitSec = sourceSnapshot.ShotDurationLimitSec
	requested.Config = sourceSnapshot.Config
	return requested, nil
}

func (s *MySQLStore) AdmitGeneration(ctx context.Context, req GenerationRequest) (AdmissionResult, error) {
	snapshot, err := normalizeGenerationRequest(req)
	if err != nil {
		return AdmissionResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return AdmissionResult{}, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, req.BatchProjectID, false)
	if err != nil {
		return AdmissionResult{}, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE batch_project_id=? AND idempotency_key=? FOR UPDATE`, req.BatchProjectID, req.RequestID))
	created := errors.Is(err, sql.ErrNoRows)
	if err != nil && !created {
		return AdmissionResult{}, err
	}
	if !created {
		// The historical unique index is case/accent insensitive. Its hit does not
		// authorize replay with a different original sequence of key bytes.
		if r.IdempotencyKey != req.RequestID || r.RunKind != "generation" {
			return AdmissionResult{}, ErrIdempotencyConflict
		}
		frozen, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
		if err != nil {
			return AdmissionResult{}, ErrIdempotencyConflict
		}
		if r.RequestedByUserID != req.RequestedByUserID {
			return AdmissionResult{}, ErrIdempotencyConflict
		}
		if snapshot.Action == GenerationActionStageRetry {
			if frozen.Action != GenerationActionStageRetry || !reflect.DeepEqual(snapshot.BookIDs, frozen.BookIDs) || snapshot.RetryStage != frozen.RetryStage || snapshot.SourceBookRunID != frozen.SourceBookRunID {
				return AdmissionResult{}, ErrIdempotencyConflict
			}
			snapshot = frozen
		} else if snapshot.SelectionAll {
			snapshot.BookIDs = append([]int64{}, frozen.BookIDs...)
		}
		if snapshot.Action != GenerationActionStageRetry {
			if frozen.SchemaVersion == 1 {
				snapshot.SchemaVersion = 1
			}
			snapshot.Config = frozen.Config
			_, hash, err := encodeGenerationSnapshot(snapshot)
			if err != nil {
				return AdmissionResult{}, err
			}
			if hash != r.RequestHash {
				return AdmissionResult{}, ErrIdempotencyConflict
			}
		}
	}
	if created {
		if snapshot.Action == GenerationActionStageRetry {
			snapshot, err = loadStageRetrySnapshot(ctx, tx, snapshot)
			if err != nil {
				return AdmissionResult{}, err
			}
		} else if snapshot.SelectionAll {
			snapshot.BookIDs, err = projectBookIDs(ctx, tx, intakeID)
			if err != nil {
				return AdmissionResult{}, err
			}
		}
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		return AdmissionResult{}, err
	}
	if created && snapshot.Action != GenerationActionStageRetry {
		snapshot.Config, err = loadGenerationConfig(ctx, tx, req.BatchProjectID)
		if err != nil {
			return AdmissionResult{}, err
		}
	}
	if created {
		body, hash, err := encodeGenerationSnapshot(snapshot)
		if err != nil {
			return AdmissionResult{}, err
		}
		r = RunRecord{BatchProjectID: req.BatchProjectID, IdempotencyKey: req.RequestID, RunAt: time.Now(), Status: RunRunning, MaxAttempts: 3, RunKind: "generation", RequestSchemaVersion: snapshot.SchemaVersion, RequestSnapshot: body, RequestHash: hash, RequestedByUserID: req.RequestedByUserID}
		if !snapshot.SelectionAll && len(snapshot.BookIDs) == 1 {
			target := snapshot.BookIDs[0]
			r.TargetBookID = &target
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO runs (batch_project_id,idempotency_key,run_at,status,max_attempts,run_kind,target_book_id,request_schema_version,request_snapshot,request_hash,requested_by_user_id,started_at) VALUES (?,?,?,'running',?,'generation',?,?,?,?,?,?)`, r.BatchProjectID, r.IdempotencyKey, r.RunAt, r.MaxAttempts, r.TargetBookID, r.RequestSchemaVersion, body, hash, r.RequestedByUserID, r.RunAt)
		if err != nil {
			return AdmissionResult{}, err
		}
		r.ID, err = res.LastInsertId()
		if err != nil {
			return AdmissionResult{}, err
		}
	}
	// Replays never repair terminal/cancelled work or change its lifecycle.
	var items []WorkItem
	if created {
		items, err = materializeGenerationBooks(ctx, tx, r, snapshot)
	} else {
		items, err = queuedGenerationItems(ctx, tx, r.ID)
	}
	if err != nil {
		return AdmissionResult{}, err
	}
	taskIDs := make([]int64, 0, len(items))
	if created {
		for _, item := range items {
			taskIDs = append(taskIDs, item.BookRunID)
		}
	} else {
		rows, err := tx.QueryContext(ctx, `SELECT id FROM book_runs WHERE run_id=? AND attempt=1 ORDER BY book_id`, r.ID)
		if err != nil {
			return AdmissionResult{}, err
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return AdmissionResult{}, err
			}
			taskIDs = append(taskIDs, id)
		}
		if err := rows.Close(); err != nil {
			return AdmissionResult{}, err
		}
		if err := rows.Err(); err != nil {
			return AdmissionResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return AdmissionResult{}, err
	}
	return AdmissionResult{Run: r, Items: items, TaskIDs: taskIDs, Created: created, Dispatch: "pending"}, nil
}

// Select candidates without row locks, then always lock Project -> Run ->
// BookRun. A cursor passes invalid or archived candidates without starving a
// later valid run, and SKIP LOCKED avoids waiting on another project's work.
func (s *MySQLStore) ClaimAndMaterializeDueRun(ctx context.Context, now time.Time) (RunClaim, []WorkItem, bool, error) {
	var cursorAt time.Time
	var cursorID int64
	for {
		var id, projectID int64
		var at time.Time
		query := `SELECT id,batch_project_id,run_at FROM runs WHERE BINARY run_kind='generation' AND status='pending' AND run_at<=?`
		args := []any{now}
		if cursorID > 0 {
			query += ` AND (run_at>? OR (run_at=? AND id>?))`
			args = append(args, cursorAt, cursorAt, cursorID)
		}
		query += ` ORDER BY run_at,id LIMIT 1`
		err := s.db.QueryRowContext(ctx, query, args...).Scan(&id, &projectID, &at)
		if errors.Is(err, sql.ErrNoRows) {
			return RunClaim{}, nil, false, nil
		}
		if err != nil {
			return RunClaim{}, nil, false, err
		}
		cursorAt, cursorID = at, id
		claim, items, ok, err := s.claimGenerationCandidate(ctx, id, projectID, now)
		if err != nil {
			return RunClaim{}, nil, false, err
		}
		if ok {
			return claim, items, true, nil
		}
	}
}
func (s *MySQLStore) claimGenerationCandidate(ctx context.Context, id, projectID int64, now time.Time) (RunClaim, []WorkItem, bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return RunClaim{}, nil, false, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, projectID, true)
	if errors.Is(err, ErrProjectArchived) || errors.Is(err, sql.ErrNoRows) {
		return RunClaim{}, nil, false, nil
	}
	if err != nil {
		return RunClaim{}, nil, false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE SKIP LOCKED`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return RunClaim{}, nil, false, nil
	}
	if err != nil {
		return RunClaim{}, nil, false, err
	}
	if r.BatchProjectID != projectID || r.RunKind != "generation" || r.Status != RunPending || r.RunAt.After(now) {
		return RunClaim{}, nil, false, nil
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if errors.Is(err, ErrInvalidGenerationRequest) {
		return RunClaim{}, nil, false, nil
	}
	if err != nil {
		return RunClaim{}, nil, false, err
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		if errors.Is(err, ErrInvalidGenerationRequest) {
			return RunClaim{}, nil, false, nil
		}
		return RunClaim{}, nil, false, err
	}
	items, err := materializeGenerationBooks(ctx, tx, r, snapshot)
	if err != nil {
		return RunClaim{}, nil, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='running',started_at=COALESCE(started_at,?) WHERE id=? AND status='pending'`, now, id); err != nil {
		return RunClaim{}, nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return RunClaim{}, nil, false, err
	}
	return RunClaim{RunID: id}, items, true, nil
}

func (s *MySQLStore) RecoverUnmaterializedRuns(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id,r.batch_project_id FROM runs r WHERE BINARY r.run_kind='generation' AND r.status='running' AND NOT EXISTS (SELECT 1 FROM book_runs br WHERE br.run_id=r.id) ORDER BY r.id`)
	if err != nil {
		return 0, err
	}
	type candidate struct{ id, projectID int64 }
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.projectID); err != nil {
			rows.Close()
			return 0, err
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	count := 0
	for _, c := range candidates {
		ok, err := s.recoverUnmaterializedRun(ctx, c.id, c.projectID, now)
		if err != nil {
			return count, err
		}
		if ok {
			count++
			if count >= limit {
				break
			}
		}
	}
	return count, nil
}
func (s *MySQLStore) recoverUnmaterializedRun(ctx context.Context, id, projectID int64, now time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, projectID, true)
	if errors.Is(err, ErrProjectArchived) || errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE SKIP LOCKED`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if r.BatchProjectID != projectID || r.RunKind != "generation" || r.Status != RunRunning {
		return false, nil
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if errors.Is(err, ErrInvalidGenerationRequest) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		if errors.Is(err, ErrInvalidGenerationRequest) {
			return false, nil
		}
		return false, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM book_runs WHERE run_id=?`, id).Scan(&count); err != nil {
		return false, err
	}
	if count != 0 {
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='pending',started_at=NULL WHERE id=? AND status='running'`, id); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
