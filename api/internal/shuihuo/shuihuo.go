package shuihuo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AssetType string

const (
	AssetReferenceImage AssetType = "reference_image"
	AssetImage          AssetType = "image"
	AssetAudio          AssetType = "audio"
	AssetVideo          AssetType = "video"
)

type AssetStatus string

const (
	AssetReady   AssetStatus = "ready"
	AssetPending AssetStatus = "pending"
	AssetFailed  AssetStatus = "failed"
)

type MediaKind string

const (
	MediaImage MediaKind = "image"
	MediaAudio MediaKind = "audio"
	MediaVideo MediaKind = "video"
)

type MediaTaskStatus string

const (
	MediaPendingExecutor MediaTaskStatus = "pending_executor"
	MediaQueued          MediaTaskStatus = "queued"
	MediaRunning         MediaTaskStatus = "running"
	MediaSucceeded       MediaTaskStatus = "succeeded"
	MediaFailed          MediaTaskStatus = "failed"
	MediaRetryableFailed MediaTaskStatus = "retryable_failed"
)

var (
	ErrConflict       = errors.New("shuihuo: version conflict")
	ErrNotFound       = errors.New("shuihuo: not found")
	ErrInvalidReorder = errors.New("shuihuo: reorder must contain every segment exactly once")
)

type Segment struct {
	ID, BatchProjectID, BookID int64
	Position, Version          int
	Text, EditRevision         string
	CreatedAt, UpdatedAt       time.Time
}
type Asset struct {
	ID, BatchProjectID, BookID, SegmentID int64
	Type                                  AssetType
	Bucket, ObjectKey                     string
	Metadata                              []byte
	Status                                AssetStatus
	CreatedAt, UpdatedAt                  time.Time
}
type MediaTask struct {
	ID, BatchProjectID, BookID, SegmentID, SourceAssetID, ProductionTaskID int64
	Kind                                                                   MediaKind
	Status                                                                 MediaTaskStatus
	Provider, Model, RequestID, ErrorCode, ErrorMessage                    string
	CreatedAt, UpdatedAt                                                   time.Time
}
type Candidate struct {
	ID, MediaTaskID, AssetID int64
	Position                 int
	Selected                 bool
	CreatedAt                time.Time
}
type CreateSegmentInput struct {
	BatchProjectID, BookID int64
	Position               int
	Text, EditRevision     string
}
type UpdateSegmentInput struct {
	Text, EditRevision string
	Version            int
}
type CreateAssetInput struct {
	BatchProjectID, BookID, SegmentID int64
	Type                              AssetType
	Bucket, ObjectKey                 string
	Metadata                          []byte
	Status                            AssetStatus
}
type CreateMediaTaskInput struct {
	BatchProjectID, BookID, SegmentID, SourceAssetID, ProductionTaskID int64
	Kind                                                               MediaKind
	Provider, Model, RequestID                                         string
}
type Store interface {
	CreateSegment(context.Context, CreateSegmentInput) (Segment, error)
	UpdateSegment(context.Context, int64, int64, int64, UpdateSegmentInput) (Segment, error)
	ListSegments(context.Context, int64, int64) ([]Segment, error)
	CreateAsset(context.Context, CreateAssetInput) (Asset, error)
	ListAssets(context.Context, int64, int64, int64) ([]Asset, error)
	CreateMediaTask(context.Context, CreateMediaTaskInput) (MediaTask, error)
	ListMediaTasks(context.Context, int64, int64) ([]MediaTask, error)
	ReorderSegments(context.Context, int64, int64, []int64) ([]Segment, error)
	ListCandidates(context.Context, int64, int64, int64) ([]Candidate, error)
	SelectCandidate(context.Context, int64, int64, int64, int64) (Candidate, error)
	RetryMediaTask(context.Context, int64, int64, int64) (MediaTask, error)
}
type Service struct{ store Store }

func NewService(s Store) *Service { return &Service{s} }
func (s *Service) CreateSegment(c context.Context, i CreateSegmentInput) (Segment, error) {
	if s == nil || s.store == nil || i.BatchProjectID <= 0 || i.BookID <= 0 || i.Position < 0 || strings.TrimSpace(i.Text) == "" {
		return Segment{}, fmt.Errorf("shuihuo: invalid segment")
	}
	return s.store.CreateSegment(c, i)
}
func (s *Service) UpdateSegment(c context.Context, p, b, id int64, i UpdateSegmentInput) (Segment, error) {
	if id <= 0 || p <= 0 || b <= 0 || i.Version <= 0 || strings.TrimSpace(i.Text) == "" {
		return Segment{}, fmt.Errorf("shuihuo: invalid segment update")
	}
	return s.store.UpdateSegment(c, p, b, id, i)
}
func (s *Service) ListSegments(c context.Context, p, b int64) ([]Segment, error) {
	return s.store.ListSegments(c, p, b)
}
func (s *Service) CreateAsset(c context.Context, i CreateAssetInput) (Asset, error) {
	if i.BatchProjectID <= 0 || i.BookID <= 0 || i.Type == "" || strings.TrimSpace(i.ObjectKey) == "" {
		return Asset{}, fmt.Errorf("shuihuo: invalid asset")
	}
	if i.Status == "" {
		i.Status = AssetReady
	}
	return s.store.CreateAsset(c, i)
}
func (s *Service) ListAssets(c context.Context, p, b, seg int64) ([]Asset, error) {
	return s.store.ListAssets(c, p, b, seg)
}
func (s *Service) CreateMediaTask(c context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	if i.BatchProjectID <= 0 || i.BookID <= 0 || i.Kind == "" || (i.Kind == MediaVideo && i.ProductionTaskID <= 0) {
		return MediaTask{}, fmt.Errorf("shuihuo: invalid media task")
	}
	return s.store.CreateMediaTask(c, i)
}
func (s *Service) ListMediaTasks(c context.Context, p, b int64) ([]MediaTask, error) {
	return s.store.ListMediaTasks(c, p, b)
}
func (s *Service) ReorderSegments(c context.Context, p, b int64, ids []int64) ([]Segment, error) {
	if p <= 0 || b <= 0 || len(ids) == 0 {
		return nil, ErrInvalidReorder
	}
	return s.store.ReorderSegments(c, p, b, ids)
}
func (s *Service) ListCandidates(c context.Context, p, b, task int64) ([]Candidate, error) {
	if p <= 0 || b <= 0 || task <= 0 {
		return nil, ErrNotFound
	}
	return s.store.ListCandidates(c, p, b, task)
}
func (s *Service) SelectCandidate(c context.Context, p, b, task, candidate int64) (Candidate, error) {
	if p <= 0 || b <= 0 || task <= 0 || candidate <= 0 {
		return Candidate{}, ErrNotFound
	}
	return s.store.SelectCandidate(c, p, b, task, candidate)
}
func (s *Service) RetryMediaTask(c context.Context, p, b, task int64) (MediaTask, error) {
	if p <= 0 || b <= 0 || task <= 0 {
		return MediaTask{}, ErrNotFound
	}
	return s.store.RetryMediaTask(c, p, b, task)
}

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db} }
func n(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}
func (s *MySQLStore) CreateSegment(c context.Context, i CreateSegmentInput) (Segment, error) {
	r, e := s.db.ExecContext(c, `INSERT INTO shuihuo_storyboard_segments(batch_project_id,book_id,position,text,edit_revision) SELECT ?,?,?,?,? FROM batch_projects p JOIN books b ON b.intake_id=p.intake_id WHERE p.id=? AND b.id=?`, i.BatchProjectID, i.BookID, i.Position, i.Text, i.EditRevision, i.BatchProjectID, i.BookID)
	if e != nil {
		return Segment{}, e
	}
	id, _ := r.LastInsertId()
	return s.segment(c, id)
}
func (s *MySQLStore) segment(c context.Context, id int64) (v Segment, e error) {
	e = s.db.QueryRowContext(c, `SELECT id,batch_project_id,book_id,position,text,edit_revision,version,created_at,updated_at FROM shuihuo_storyboard_segments WHERE id=?`, id).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &v.Position, &v.Text, &v.EditRevision, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	return
}
func (s *MySQLStore) UpdateSegment(c context.Context, p, b, id int64, i UpdateSegmentInput) (Segment, error) {
	r, e := s.db.ExecContext(c, `UPDATE shuihuo_storyboard_segments SET text=?,edit_revision=?,version=version+1 WHERE id=? AND batch_project_id=? AND book_id=? AND version=?`, i.Text, i.EditRevision, id, p, b, i.Version)
	if e != nil {
		return Segment{}, e
	}
	x, _ := r.RowsAffected()
	if x == 0 {
		return Segment{}, ErrConflict
	}
	return s.segment(c, id)
}
func (s *MySQLStore) ReorderSegments(c context.Context, p, b int64, ids []int64) ([]Segment, error) {
	tx, e := s.db.BeginTx(c, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	rows, e := tx.QueryContext(c, `SELECT id FROM shuihuo_storyboard_segments WHERE batch_project_id=? AND book_id=? ORDER BY position,id FOR UPDATE`, p, b)
	if e != nil {
		return nil, e
	}
	var existing []int64
	for rows.Next() {
		var id int64
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return nil, e
		}
		existing = append(existing, id)
	}
	rows.Close()
	if len(existing) != len(ids) {
		return nil, ErrInvalidReorder
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return nil, ErrInvalidReorder
		}
		seen[id] = true
	}
	for _, id := range existing {
		if !seen[id] {
			return nil, ErrInvalidReorder
		}
	}
	if _, e = tx.ExecContext(c, `UPDATE shuihuo_storyboard_segments SET position=position+1000000 WHERE batch_project_id=? AND book_id=?`, p, b); e != nil {
		return nil, e
	}
	for pos, id := range ids {
		if _, e = tx.ExecContext(c, `UPDATE shuihuo_storyboard_segments SET position=?,version=version+1 WHERE id=? AND batch_project_id=? AND book_id=?`, pos, id, p, b); e != nil {
			return nil, e
		}
	}
	if e = tx.Commit(); e != nil {
		return nil, e
	}
	return s.ListSegments(c, p, b)
}
func (s *MySQLStore) ListCandidates(c context.Context, p, b, taskID int64) (out []Candidate, e error) {
	rows, e := s.db.QueryContext(c, `SELECT c.id,c.media_task_id,c.asset_id,c.position,c.selected,c.created_at FROM shuihuo_media_candidates c JOIN shuihuo_media_tasks t ON t.id=c.media_task_id WHERE c.media_task_id=? AND t.batch_project_id=? AND t.book_id=? ORDER BY c.position,c.id`, taskID, p, b)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v Candidate
		if e = rows.Scan(&v.ID, &v.MediaTaskID, &v.AssetID, &v.Position, &v.Selected, &v.CreatedAt); e != nil {
			return
		}
		out = append(out, v)
	}
	e = rows.Err()
	return
}
func (s *MySQLStore) SelectCandidate(c context.Context, p, b, taskID, candidateID int64) (out Candidate, e error) {
	tx, e := s.db.BeginTx(c, nil)
	if e != nil {
		return out, e
	}
	defer tx.Rollback()
	var id int64
	e = tx.QueryRowContext(c, `SELECT c.id FROM shuihuo_media_candidates c JOIN shuihuo_media_tasks t ON t.id=c.media_task_id WHERE c.id=? AND c.media_task_id=? AND t.batch_project_id=? AND t.book_id=? FOR UPDATE`, candidateID, taskID, p, b).Scan(&id)
	if errors.Is(e, sql.ErrNoRows) {
		return out, ErrNotFound
	}
	if e != nil {
		return out, e
	}
	if _, e = tx.ExecContext(c, `UPDATE shuihuo_media_candidates SET selected=0 WHERE media_task_id=?`, taskID); e != nil {
		return out, e
	}
	if _, e = tx.ExecContext(c, `UPDATE shuihuo_media_candidates SET selected=1 WHERE id=? AND media_task_id=?`, candidateID, taskID); e != nil {
		return out, e
	}
	if e = tx.Commit(); e != nil {
		return out, e
	}
	e = s.db.QueryRowContext(c, `SELECT id,media_task_id,asset_id,position,selected,created_at FROM shuihuo_media_candidates WHERE id=?`, candidateID).Scan(&out.ID, &out.MediaTaskID, &out.AssetID, &out.Position, &out.Selected, &out.CreatedAt)
	return out, e
}
func (s *MySQLStore) RetryMediaTask(c context.Context, p, b, taskID int64) (out MediaTask, e error) {
	r, e := s.db.ExecContext(c, `UPDATE shuihuo_media_tasks SET status='pending_executor',error_code='',error_message='' WHERE id=? AND batch_project_id=? AND book_id=? AND production_task_id IS NULL AND status IN ('pending_executor','retryable_failed')`, taskID, p, b)
	if e != nil {
		return out, e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return out, ErrNotFound
	}
	return s.task(c, taskID)
}
func (s *MySQLStore) ListSegments(c context.Context, p, b int64) (out []Segment, e error) {
	rows, e := s.db.QueryContext(c, `SELECT id,batch_project_id,book_id,position,text,edit_revision,version,created_at,updated_at FROM shuihuo_storyboard_segments WHERE batch_project_id=? AND book_id=? ORDER BY position,id`, p, b)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v Segment
		if e = rows.Scan(&v.ID, &v.BatchProjectID, &v.BookID, &v.Position, &v.Text, &v.EditRevision, &v.Version, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return
		}
		out = append(out, v)
	}
	e = rows.Err()
	return
}
func (s *MySQLStore) CreateAsset(c context.Context, i CreateAssetInput) (Asset, error) {
	m := i.Metadata
	if len(m) == 0 {
		m = []byte(`{}`)
	}
	r, e := s.db.ExecContext(c, `INSERT INTO shuihuo_media_assets(batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status) VALUES(?,?,?,?,?,?,?,?)`, i.BatchProjectID, i.BookID, n(i.SegmentID), i.Type, i.Bucket, i.ObjectKey, m, i.Status)
	if e != nil {
		return Asset{}, e
	}
	id, _ := r.LastInsertId()
	return s.asset(c, id)
}
func (s *MySQLStore) asset(c context.Context, id int64) (v Asset, e error) {
	var q sql.NullInt64
	e = s.db.QueryRowContext(c, `SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE id=?`, id).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &q, &v.Type, &v.Bucket, &v.ObjectKey, &v.Metadata, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	v.SegmentID = q.Int64
	return
}
func (s *MySQLStore) ListAssets(c context.Context, p, b, seg int64) (out []Asset, e error) {
	q := `SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE batch_project_id=? AND book_id=?`
	a := []any{p, b}
	if seg > 0 {
		q += ` AND segment_id=?`
		a = append(a, seg)
	}
	rows, e := s.db.QueryContext(c, q, a...)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v Asset
		var x sql.NullInt64
		if e = rows.Scan(&v.ID, &v.BatchProjectID, &v.BookID, &x, &v.Type, &v.Bucket, &v.ObjectKey, &v.Metadata, &v.Status, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return
		}
		v.SegmentID = x.Int64
		out = append(out, v)
	}
	e = rows.Err()
	return
}
func (s *MySQLStore) CreateMediaTask(c context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	r, e := s.db.ExecContext(c, `INSERT INTO shuihuo_media_tasks(batch_project_id,book_id,segment_id,source_asset_id,production_task_id,media_kind,provider,model,request_id)VALUES(?,?,?,?,?,?,?,?,?)`, i.BatchProjectID, i.BookID, n(i.SegmentID), n(i.SourceAssetID), n(i.ProductionTaskID), i.Kind, i.Provider, i.Model, i.RequestID)
	if e != nil {
		return MediaTask{}, e
	}
	id, _ := r.LastInsertId()
	return s.task(c, id)
}
func (s *MySQLStore) task(c context.Context, id int64) (v MediaTask, e error) {
	var a, b, d sql.NullInt64
	e = s.db.QueryRowContext(c, `SELECT id,batch_project_id,book_id,segment_id,source_asset_id,production_task_id,media_kind,status,provider,model,request_id,error_code,error_message,created_at,updated_at FROM shuihuo_media_tasks WHERE id=?`, id).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &a, &b, &d, &v.Kind, &v.Status, &v.Provider, &v.Model, &v.RequestID, &v.ErrorCode, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt)
	v.SegmentID, v.SourceAssetID, v.ProductionTaskID = a.Int64, b.Int64, d.Int64
	return
}
func (s *MySQLStore) ListMediaTasks(c context.Context, p, b int64) (out []MediaTask, e error) {
	rows, e := s.db.QueryContext(c, `SELECT t.id,t.batch_project_id,t.book_id,t.segment_id,t.source_asset_id,t.production_task_id,t.media_kind,CASE WHEN t.production_task_id IS NULL THEN t.status ELSE vt.status END,t.provider,t.model,t.request_id,CASE WHEN t.production_task_id IS NULL THEN t.error_code ELSE vt.error_code END,CASE WHEN t.production_task_id IS NULL THEN t.error_message ELSE vt.error_message END,t.created_at,t.updated_at FROM shuihuo_media_tasks t LEFT JOIN video_production_tasks vt ON vt.id=t.production_task_id WHERE t.batch_project_id=? AND t.book_id=? ORDER BY t.id`, p, b)
	if e != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v MediaTask
		var a, x, d sql.NullInt64
		if e = rows.Scan(&v.ID, &v.BatchProjectID, &v.BookID, &a, &x, &d, &v.Kind, &v.Status, &v.Provider, &v.Model, &v.RequestID, &v.ErrorCode, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return
		}
		v.SegmentID, v.SourceAssetID, v.ProductionTaskID = a.Int64, x.Int64, d.Int64
		out = append(out, v)
	}
	e = rows.Err()
	return
}
