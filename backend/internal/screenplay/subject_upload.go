package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/subjectupload"
	"sync"
)

// SubjectUploadOperation is private persistence; HTTP must expose an explicit DTO.
type SubjectUploadOperation struct {
	Owner, RequestKey, Digest, Status, Phase, Message string
	Revision                                          int64
	Source                                            agentworkspace.Asset
	SourceObjectKey                                   string
	Progress                                          subjectupload.Progress
	Result                                            *subjectupload.Result
}
type SubjectUploadRepository interface {
	GetSubjectUpload(context.Context, string, string) (SubjectUploadOperation, error)
	SaveSubjectUpload(context.Context, SubjectUploadOperation, int64) (SubjectUploadOperation, error)
}
type MemorySubjectUploadRepository struct {
	mu   sync.Mutex
	rows map[string]SubjectUploadOperation
}

func NewMemorySubjectUploadRepository() *MemorySubjectUploadRepository {
	return &MemorySubjectUploadRepository{rows: map[string]SubjectUploadOperation{}}
}
func cloneSubjectUpload(v SubjectUploadOperation) SubjectUploadOperation {
	if v.Result != nil {
		r := *v.Result
		v.Result = &r
	}
	return v
}
func (m *MemorySubjectUploadRepository) GetSubjectUpload(_ context.Context, o, k string) (SubjectUploadOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.rows[o+"\x00"+k]
	if !ok {
		return v, ErrNotFound
	}
	return cloneSubjectUpload(v), nil
}
func (m *MemorySubjectUploadRepository) SaveSubjectUpload(_ context.Context, v SubjectUploadOperation, rev int64) (SubjectUploadOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := v.Owner + "\x00" + v.RequestKey
	if v.Owner == "" || v.RequestKey == "" || v.Digest == "" {
		return v, ErrInvalidArgument
	}
	old, exists := m.rows[key]
	if exists && old.Revision != rev || !exists && rev != 0 || exists && rev == 0 {
		return v, ErrRevisionConflict
	}
	v.Revision = rev + 1
	m.rows[key] = cloneSubjectUpload(v)
	return cloneSubjectUpload(v), nil
}
func (s *MySQLStore) GetSubjectUpload(ctx context.Context, o, k string) (SubjectUploadOperation, error) {
	var raw []byte
	var rev int64
	e := s.db.QueryRowContext(ctx, "SELECT operation_json,revision FROM screenplay_subject_uploads WHERE owner_username=? AND request_key=?", o, k).Scan(&raw, &rev)
	if errors.Is(e, sql.ErrNoRows) {
		return SubjectUploadOperation{}, ErrNotFound
	}
	if e != nil {
		return SubjectUploadOperation{}, e
	}
	var v SubjectUploadOperation
	e = json.Unmarshal(raw, &v)
	v.Source.ObjectKey = v.SourceObjectKey
	v.Revision = rev
	return v, e
}
func (s *MySQLStore) SaveSubjectUpload(ctx context.Context, v SubjectUploadOperation, rev int64) (SubjectUploadOperation, error) {
	if v.Owner == "" || v.RequestKey == "" || v.Digest == "" {
		return v, ErrInvalidArgument
	}
	v.Revision = rev + 1
	v.SourceObjectKey = v.Source.ObjectKey
	raw, e := json.Marshal(v)
	if e != nil {
		return v, e
	}
	if rev == 0 {
		_, e = s.db.ExecContext(ctx, "INSERT INTO screenplay_subject_uploads(owner_username,request_key,revision,operation_json) VALUES(?,?,?,?)", v.Owner, v.RequestKey, v.Revision, raw)
		if e != nil {
			if _, found := s.GetSubjectUpload(ctx, v.Owner, v.RequestKey); found == nil {
				return v, ErrRevisionConflict
			}
			return v, e
		}
		return v, nil
	}
	res, e := s.db.ExecContext(ctx, "UPDATE screenplay_subject_uploads SET revision=?,operation_json=? WHERE owner_username=? AND request_key=? AND revision=?", v.Revision, raw, v.Owner, v.RequestKey, rev)
	if e != nil {
		return v, e
	}
	n, e := res.RowsAffected()
	if e == nil && n != 1 {
		e = ErrRevisionConflict
	}
	return v, e
}
