package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

// ContentRecord persists typed document, selection, style-source and continuity
// records. Project transactions provide CAS; callers cannot supply object keys.
type ContentRecord struct {
	Owner      string `json:"-"`
	ProjectID  string
	Kind       string
	ID         string
	Status     string
	LeaseUntil time.Time
	Data       json.RawMessage
}
type ContentStore interface {
	GetContentRecord(context.Context, string, string, string, string) (ContentRecord, bool, error)
	PutContentRecord(context.Context, ContentRecord) error
	ListContentRecords(context.Context, string, string, string) ([]ContentRecord, error)
	ListParseDispatch(context.Context, int) ([]ContentRecord, error)
}

func (s *MemoryStore) GetContentRecord(ctx context.Context, owner, project, kind, id string) (ContentRecord, bool, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return ContentRecord{}, false, e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.contentRecords[memoryKey(owner, project, kind, id)]
	r.Data = cloneRawJSON(r.Data)
	return r, ok, nil
}
func (s *MemoryStore) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if _, e := s.GetProject(ctx, r.Owner, r.ProjectID); e != nil {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r.Data = cloneRawJSON(r.Data)
	s.contentRecords[memoryKey(r.Owner, r.ProjectID, r.Kind, r.ID)] = r
	return nil
}
func (s *MemoryStore) ListContentRecords(ctx context.Context, owner, project, kind string) ([]ContentRecord, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return nil, e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []ContentRecord{}
	for _, r := range s.contentRecords {
		if r.Owner == owner && r.ProjectID == project && r.Kind == kind {
			r.Data = cloneRawJSON(r.Data)
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (s *MemoryStore) ListParseDispatch(_ context.Context, limit int) ([]ContentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []ContentRecord{}
	for _, r := range s.contentRecords {
		if r.Kind == "document" && (r.Status == "queued" || r.Status == "parsing") && !r.LeaseUntil.After(time.Now().UTC()) && !projectReadOnly(s.projects[memoryKey(r.Owner, r.ProjectID)].Status) {
			r.Data = cloneRawJSON(r.Data)
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (s *MySQLStore) GetContentRecord(ctx context.Context, owner, project, kind, id string) (ContentRecord, bool, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return ContentRecord{}, false, e
	}
	r := ContentRecord{Owner: owner, ProjectID: project, Kind: kind, ID: id}
	var until sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT status,lease_until,snapshot_json FROM novel_content_records WHERE owner_username=? AND project_id=? AND kind=? AND record_id=?`, owner, project, kind, id).Scan(&r.Status, &until, &r.Data)
	if errors.Is(err, sql.ErrNoRows) {
		return r, false, nil
	}
	r.LeaseUntil = until.Time
	return r, err == nil, persistenceError(err)
}
func (s *MySQLStore) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if !json.Valid(r.Data) {
		return persistenceError(errors.New("invalid record JSON"))
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO novel_content_records (owner_username,project_id,kind,record_id,status,lease_until,snapshot_json,updated_at) VALUES (?,?,?,?,?,?,?,UTC_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE status=VALUES(status),lease_until=VALUES(lease_until),snapshot_json=VALUES(snapshot_json),updated_at=VALUES(updated_at)`, r.Owner, r.ProjectID, r.Kind, r.ID, r.Status, nullableLease(r.LeaseUntil), []byte(r.Data))
	return persistenceError(err)
}
func (s *MySQLStore) ListContentRecords(ctx context.Context, owner, project, kind string) ([]ContentRecord, error) {
	if _, e := s.GetProject(ctx, owner, project); e != nil {
		return nil, e
	}
	return s.readContentRecords(ctx, `SELECT owner_username,project_id,kind,record_id,status,lease_until,snapshot_json FROM novel_content_records WHERE owner_username=? AND project_id=? AND kind=? ORDER BY record_id`, owner, project, kind)
}
func (s *MySQLStore) ListParseDispatch(ctx context.Context, limit int) ([]ContentRecord, error) {
	return s.readContentRecords(ctx, `SELECT c.owner_username,c.project_id,c.kind,c.record_id,c.status,c.lease_until,c.snapshot_json FROM novel_content_records c JOIN novel_projects p ON p.owner_username=c.owner_username AND p.project_id=c.project_id WHERE c.kind='document' AND c.status IN ('queued','parsing') AND (c.lease_until IS NULL OR c.lease_until<=UTC_TIMESTAMP(6)) AND p.deleted_at IS NULL AND p.status NOT IN ('archived','purging') ORDER BY c.updated_at LIMIT ?`, limit)
}
func (s *MySQLStore) readContentRecords(ctx context.Context, q string, args ...any) ([]ContentRecord, error) {
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, persistenceError(e)
	}
	defer rows.Close()
	out := []ContentRecord{}
	for rows.Next() {
		var r ContentRecord
		var until sql.NullTime
		if e = rows.Scan(&r.Owner, &r.ProjectID, &r.Kind, &r.ID, &r.Status, &until, &r.Data); e != nil {
			return nil, persistenceError(e)
		}
		r.LeaseUntil = until.Time
		out = append(out, r)
	}
	return out, persistenceError(rows.Err())
}
func putContent(ctx context.Context, st Store, owner, project, kind, id, status string, until time.Time, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return st.PutContentRecord(ctx, ContentRecord{Owner: owner, ProjectID: project, Kind: kind, ID: id, Status: status, LeaseUntil: until, Data: b})
}
func readContent(ctx context.Context, st Store, owner, project, kind, id string, v any) (bool, error) {
	r, ok, e := st.GetContentRecord(ctx, owner, project, kind, id)
	if e != nil || !ok {
		return ok, e
	}
	if e = json.Unmarshal(r.Data, v); e != nil {
		return false, persistenceError(e)
	}
	return true, nil
}
func inputError(message string) error { return &DomainError{Code: ErrInputRequired, Message: message} }
func documentID(id string) bool       { return strings.HasPrefix(id, "docref-") }

func decodeRecord(data json.RawMessage, v any) error {
	return persistenceError(json.Unmarshal(data, v))
}
