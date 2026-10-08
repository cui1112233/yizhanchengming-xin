package shuihuo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
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
	ErrConflict           = errors.New("shuihuo: version conflict")
	ErrNotFound           = errors.New("shuihuo: not found")
	ErrInvalidReorder     = errors.New("shuihuo: reorder must contain every segment exactly once")
	ErrStorageUnavailable = errors.New("shuihuo: object storage is unavailable")
	ErrInvalidUpload      = errors.New("shuihuo: invalid upload")
)

type Segment struct {
	ID             int64     `json:"id"`
	BatchProjectID int64     `json:"batchProjectId"`
	BookID         int64     `json:"bookId"`
	Position       int       `json:"position"`
	Version        int       `json:"version"`
	Text           string    `json:"text"`
	EditRevision   string    `json:"editRevision"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type Asset struct {
	ID             int64       `json:"id"`
	BatchProjectID int64       `json:"batchProjectId"`
	BookID         int64       `json:"bookId"`
	SegmentID      int64       `json:"segmentId"`
	Type           AssetType   `json:"type"`
	Bucket         string      `json:"bucket"`
	ObjectKey      string      `json:"objectKey"`
	Metadata       []byte      `json:"metadata"`
	Status         AssetStatus `json:"status"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}
type MediaTask struct {
	ID               int64           `json:"id"`
	BatchProjectID   int64           `json:"batchProjectId"`
	BookID           int64           `json:"bookId"`
	SegmentID        int64           `json:"segmentId"`
	SourceAssetID    int64           `json:"sourceAssetId"`
	ProductionTaskID int64           `json:"productionTaskId"`
	Kind             MediaKind       `json:"kind"`
	Status           MediaTaskStatus `json:"status"`
	Provider         string          `json:"provider"`
	Model            string          `json:"model"`
	RequestID        string          `json:"requestId"`
	ErrorCode        string          `json:"errorCode"`
	ErrorMessage     string          `json:"errorMessage"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}
type Candidate struct {
	ID          int64     `json:"id"`
	MediaTaskID int64     `json:"mediaTaskId"`
	AssetID     int64     `json:"assetId"`
	Position    int       `json:"position"`
	Selected    bool      `json:"selected"`
	CreatedAt   time.Time `json:"createdAt"`
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
type ObjectStore interface {
	PutObjectFromFile(context.Context, string, string, string) error
	GetObject(context.Context, string, string) (io.ReadCloser, error)
}
type UploadAssetInput struct {
	BatchProjectID, BookID, SegmentID int64
	Type                              AssetType
	Filename, ContentType             string
	Body                              io.Reader
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
	MarkExecutorUnavailable(context.Context, int64, string, string) (MediaTask, error)
	GetAsset(context.Context, int64, int64, int64) (Asset, error)
}
type Service struct {
	store    Store
	objects  ObjectStore
	bucket   string
	provider MediaProvider
	queue    taskruntime.Queue
}

func NewService(s Store, objects ...ObjectStore) *Service {
	out := &Service{store: s}
	if len(objects) > 0 {
		out.objects = objects[0]
	}
	return out
}
func (s *Service) SetBucket(bucket string) {
	if s != nil {
		s.bucket = strings.TrimSpace(bucket)
	}
}

// SetMediaExecutor attaches the configured server-side Provider and shared
// taskruntime queue. Neither credentials nor provider configuration are exposed
// through the HTTP API.
func (s *Service) SetMediaExecutor(provider MediaProvider, queue taskruntime.Queue) {
	s.configureMediaExecutor(provider, queue)
}
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

// UploadAsset is the only browser upload path. Bucket and key are generated
// server-side, after scope and media validation, so a client cannot overwrite
// another project's TOS object.
func (s *Service) UploadAsset(c context.Context, in UploadAssetInput) (Asset, error) {
	if s == nil || s.store == nil || s.objects == nil || s.bucket == "" {
		return Asset{}, ErrStorageUnavailable
	}
	if in.BatchProjectID <= 0 || in.BookID <= 0 || in.Body == nil {
		return Asset{}, ErrInvalidUpload
	}
	contentType, extension, limit, err := uploadRules(in.Type, in.ContentType)
	if err != nil {
		return Asset{}, err
	}
	tmp, err := os.CreateTemp("", "shuihuo-upload-*")
	if err != nil {
		return Asset{}, fmt.Errorf("shuihuo: create temp file: %w", err)
	}
	path := tmp.Name()
	defer os.Remove(path)
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(tmp, hash), io.LimitReader(in.Body, limit+1))
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil || written <= 0 || written > limit {
		return Asset{}, ErrInvalidUpload
	}
	probe, err := os.Open(path)
	if err != nil {
		return Asset{}, err
	}
	header := make([]byte, 512)
	n, _ := probe.Read(header)
	_ = probe.Close()
	if !contentMatches(in.Type, contentType, http.DetectContentType(header[:n])) {
		return Asset{}, ErrInvalidUpload
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return Asset{}, err
	}
	key := fmt.Sprintf("shuihuo/project-%d/book-%d/%s.%s", in.BatchProjectID, in.BookID, hex.EncodeToString(random), extension)
	if err := s.objects.PutObjectFromFile(c, s.bucket, key, path); err != nil {
		return Asset{}, fmt.Errorf("shuihuo: upload TOS object: %w", err)
	}
	metadata := []byte(fmt.Sprintf(`{"filename":%q,"contentType":%q,"bytes":%d,"sha256":%q}`, filepath.Base(in.Filename), contentType, written, hex.EncodeToString(hash.Sum(nil))))
	return s.store.CreateAsset(c, CreateAssetInput{BatchProjectID: in.BatchProjectID, BookID: in.BookID, SegmentID: in.SegmentID, Type: in.Type, Bucket: s.bucket, ObjectKey: key, Metadata: metadata, Status: AssetReady})
}
func (s *Service) OpenAsset(c context.Context, p, b, id int64) (Asset, io.ReadCloser, error) {
	if s == nil || s.objects == nil {
		return Asset{}, nil, ErrStorageUnavailable
	}
	asset, err := s.store.GetAsset(c, p, b, id)
	if err != nil {
		return Asset{}, nil, err
	}
	if asset.Status != AssetReady || asset.Bucket == "" || asset.ObjectKey == "" {
		return Asset{}, nil, ErrNotFound
	}
	body, err := s.objects.GetObject(c, asset.Bucket, asset.ObjectKey)
	if err != nil {
		return Asset{}, nil, fmt.Errorf("shuihuo: fetch TOS object: %w", err)
	}
	return asset, body, nil
}
func uploadRules(kind AssetType, supplied string) (string, string, int64, error) {
	supplied = strings.ToLower(strings.TrimSpace(strings.Split(supplied, ";")[0]))
	switch kind {
	case AssetImage, AssetReferenceImage:
		switch supplied {
		case "image/jpeg":
			return supplied, "jpg", 20 << 20, nil
		case "image/png":
			return supplied, "png", 20 << 20, nil
		case "image/webp":
			return supplied, "webp", 20 << 20, nil
		}
	case AssetAudio:
		switch supplied {
		case "audio/mpeg":
			return supplied, "mp3", 200 << 20, nil
		case "audio/wav", "audio/x-wav":
			return "audio/wav", "wav", 200 << 20, nil
		case "audio/mp4":
			return supplied, "m4a", 200 << 20, nil
		}
	case AssetVideo:
		switch supplied {
		case "video/mp4":
			return supplied, "mp4", 2 << 30, nil
		case "video/webm":
			return supplied, "webm", 2 << 30, nil
		}
	}
	return "", "", 0, ErrInvalidUpload
}
func contentMatches(kind AssetType, supplied, detected string) bool {
	if kind == AssetImage || kind == AssetReferenceImage {
		return (supplied == "image/jpeg" && detected == "image/jpeg") || (supplied == "image/png" && detected == "image/png") || (supplied == "image/webp" && detected == "image/webp")
	}
	if kind == AssetAudio {
		return strings.HasPrefix(detected, "audio/") || detected == "application/octet-stream"
	}
	return kind == AssetVideo && (detected == "video/mp4" || detected == "application/octet-stream")
}
func (s *Service) CreateMediaTask(c context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	if i.BatchProjectID <= 0 || i.BookID <= 0 || i.Kind == "" || (i.Kind == MediaVideo && i.ProductionTaskID <= 0) {
		return MediaTask{}, fmt.Errorf("shuihuo: invalid media task")
	}
	task, err := s.store.CreateMediaTask(c, i)
	if err != nil || i.Kind == MediaVideo {
		return task, err
	}
	return s.enqueueMediaTask(c, task)
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
	item, err := s.store.RetryMediaTask(c, p, b, task)
	if err != nil || item.Kind == MediaVideo {
		return item, err
	}
	return s.enqueueMediaTask(c, item)
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
	// Both the book and optional segment must be inside this project scope.
	// This also prevents callers from attaching an uploaded object to another
	// user's storyboard by guessing a segment id.
	r, e := s.db.ExecContext(c, `INSERT INTO shuihuo_media_assets(batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status)
SELECT ?,?,?,?,?,?,?,? FROM batch_projects p JOIN books b ON b.intake_id=p.intake_id
WHERE p.id=? AND b.id=? AND (?=0 OR EXISTS(SELECT 1 FROM shuihuo_storyboard_segments s WHERE s.id=? AND s.batch_project_id=? AND s.book_id=?))`,
		i.BatchProjectID, i.BookID, n(i.SegmentID), i.Type, i.Bucket, i.ObjectKey, m, i.Status,
		i.BatchProjectID, i.BookID, i.SegmentID, i.SegmentID, i.BatchProjectID, i.BookID)
	if e != nil {
		return Asset{}, e
	}
	if affected, _ := r.RowsAffected(); affected == 0 {
		return Asset{}, ErrNotFound
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
func (s *MySQLStore) GetAsset(c context.Context, p, b, id int64) (v Asset, e error) {
	var q sql.NullInt64
	e = s.db.QueryRowContext(c, `SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE id=? AND batch_project_id=? AND book_id=?`, id, p, b).Scan(&v.ID, &v.BatchProjectID, &v.BookID, &q, &v.Type, &v.Bucket, &v.ObjectKey, &v.Metadata, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(e, sql.ErrNoRows) {
		return Asset{}, ErrNotFound
	}
	v.SegmentID = q.Int64
	return v, e
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
func (s *MySQLStore) MarkExecutorUnavailable(c context.Context, id int64, code, message string) (MediaTask, error) {
	r, e := s.db.ExecContext(c, `UPDATE shuihuo_media_tasks SET status='failed',error_code=?,error_message=? WHERE id=? AND production_task_id IS NULL AND status='pending_executor'`, code, message, id)
	if e != nil {
		return MediaTask{}, e
	}
	changed, _ := r.RowsAffected()
	if changed == 0 {
		return MediaTask{}, ErrNotFound
	}
	return s.task(c, id)
}
func (s *MySQLStore) QueueMediaTask(c context.Context, id int64) (MediaTask, error) {
	r, err := s.db.ExecContext(c, `UPDATE shuihuo_media_tasks SET status='queued',error_code='',error_message='' WHERE id=? AND production_task_id IS NULL AND status='pending_executor'`, id)
	if err != nil {
		return MediaTask{}, err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return MediaTask{}, ErrNotFound
	}
	return s.task(c, id)
}
func (s *MySQLStore) StartMediaTask(c context.Context, id int64) (bool, error) {
	r, err := s.db.ExecContext(c, `UPDATE shuihuo_media_tasks SET status='running',error_code='',error_message='' WHERE id=? AND production_task_id IS NULL AND status='queued'`, id)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n == 1, nil
}
func (s *MySQLStore) MediaTaskForExecution(c context.Context, id int64) (MediaTask, Segment, error) {
	task, err := s.task(c, id)
	if err != nil {
		return MediaTask{}, Segment{}, err
	}
	if task.ProductionTaskID != 0 || (task.Kind != MediaImage && task.Kind != MediaAudio) || task.Status != MediaRunning || task.SegmentID <= 0 {
		return MediaTask{}, Segment{}, ErrNotFound
	}
	segment, err := s.segment(c, task.SegmentID)
	if err != nil {
		return MediaTask{}, Segment{}, err
	}
	if segment.BatchProjectID != task.BatchProjectID || segment.BookID != task.BookID || strings.TrimSpace(segment.Text) == "" {
		return MediaTask{}, Segment{}, ErrNotFound
	}
	return task, segment, nil
}
func (s *MySQLStore) CompleteMediaTask(c context.Context, taskID, assetID int64) (bool, error) {
	tx, err := s.db.BeginTx(c, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var projectID, bookID int64
	err = tx.QueryRowContext(c, `SELECT batch_project_id,book_id FROM shuihuo_media_tasks WHERE id=? AND production_task_id IS NULL AND status='running' FOR UPDATE`, taskID).Scan(&projectID, &bookID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var exists int
	if err = tx.QueryRowContext(c, `SELECT 1 FROM shuihuo_media_assets WHERE id=? AND batch_project_id=? AND book_id=? AND status='ready'`, assetID, projectID, bookID).Scan(&exists); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(c, `INSERT INTO shuihuo_media_candidates(media_task_id,asset_id,position,selected) VALUES(?,?,(SELECT COALESCE(MAX(c.position),-1)+1 FROM (SELECT position FROM shuihuo_media_candidates WHERE media_task_id=?) c),1)`, taskID, assetID, taskID); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(c, `UPDATE shuihuo_media_candidates SET selected=0 WHERE media_task_id=? AND asset_id<>?`, taskID, assetID); err != nil {
		return false, err
	}
	r, err := tx.ExecContext(c, `UPDATE shuihuo_media_tasks SET status='succeeded',error_code='',error_message='' WHERE id=? AND status='running'`, taskID)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return false, nil
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
func (s *MySQLStore) FailMediaTask(c context.Context, id int64, code, message string, retryable bool) (bool, error) {
	state := MediaFailed
	if retryable {
		state = MediaRetryableFailed
	}
	r, err := s.db.ExecContext(c, `UPDATE shuihuo_media_tasks SET status=?,error_code=?,error_message=? WHERE id=? AND production_task_id IS NULL AND status IN ('queued','running')`, state, code, message, id)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n == 1, nil
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
