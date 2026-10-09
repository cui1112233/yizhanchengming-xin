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
}

type AdmissionResult struct {
	Run      RunRecord
	Items    []WorkItem
	Created  bool
	Dispatch string
}

// GenerationSnapshot is the immutable, versioned execution input. Provider
// credentials, source text, client durations and Prompt content cannot enter it.
type GenerationSnapshot struct {
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
	return GenerationSnapshot{SchemaVersion: 1, BatchProjectID: r.BatchProjectID, BookIDs: unique, SelectionAll: len(r.BookIDs) == 0, HookEnabled: r.HookEnabled, PlotMode: r.PlotMode, DirectorMode: mode, MatchAudio: r.MatchAudio, ShotDurationLimitSec: duration, RequestedByUserID: r.RequestedByUserID}, nil
}

func encodeGenerationSnapshot(s GenerationSnapshot) ([]byte, string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(b)
	return b, hex.EncodeToString(sum[:]), nil
}

func decodeGenerationSnapshot(version int, b []byte, hash string, projectID, actorID int64) (GenerationSnapshot, error) {
	var s GenerationSnapshot
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return s, ErrInvalidGenerationRequest
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return s, ErrInvalidGenerationRequest
	}
	if version != 1 || s.SchemaVersion != 1 || s.BatchProjectID != projectID || s.RequestedByUserID != actorID || len(s.BookIDs) == 0 {
		return s, ErrInvalidGenerationRequest
	}
	normalized, err := normalizeGenerationRequest(GenerationRequest{BatchProjectID: s.BatchProjectID, BookIDs: s.BookIDs, HookEnabled: s.HookEnabled, PlotMode: s.PlotMode, DirectorMode: s.DirectorMode, MatchAudio: s.MatchAudio, ShotDurationLimitSec: s.ShotDurationLimitSec, RequestID: "snapshot", RequestedByUserID: s.RequestedByUserID})
	normalized.SelectionAll = s.SelectionAll
	if err != nil || !reflect.DeepEqual(s, normalized) {
		return s, ErrInvalidGenerationRequest
	}
	_, canonicalHash, err := encodeGenerationSnapshot(s)
	if err != nil || canonicalHash != hash {
		return s, ErrInvalidGenerationRequest
	}
	return s, nil
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
		if snapshot.SelectionAll {
			snapshot.BookIDs = append([]int64{}, frozen.BookIDs...)
		}
		_, hash, err := encodeGenerationSnapshot(snapshot)
		if err != nil {
			return AdmissionResult{}, err
		}
		if hash != r.RequestHash {
			return AdmissionResult{}, ErrIdempotencyConflict
		}
	}
	if created && snapshot.SelectionAll {
		snapshot.BookIDs, err = projectBookIDs(ctx, tx, intakeID)
		if err != nil {
			return AdmissionResult{}, err
		}
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		return AdmissionResult{}, err
	}
	if created {
		body, hash, err := encodeGenerationSnapshot(snapshot)
		if err != nil {
			return AdmissionResult{}, err
		}
		r = RunRecord{BatchProjectID: req.BatchProjectID, IdempotencyKey: req.RequestID, RunAt: time.Now(), Status: RunRunning, MaxAttempts: 3, RunKind: "generation", RequestSchemaVersion: 1, RequestSnapshot: body, RequestHash: hash, RequestedByUserID: req.RequestedByUserID}
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
	if err := tx.Commit(); err != nil {
		return AdmissionResult{}, err
	}
	return AdmissionResult{Run: r, Items: items, Created: created, Dispatch: "pending"}, nil
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
