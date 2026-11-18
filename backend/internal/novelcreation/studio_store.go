package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

func studioMissing() error {
	return &DomainError{Code: ErrSessionNotFound, Message: "创作会话不存在"}
}
func (s *MemoryStore) ReserveStudioCreation(_ context.Context, owner, key, digest string, raw json.RawMessage) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.studios {
		if !strings.HasPrefix(k, owner+"\x00") {
			continue
		}
		var r studioRecord
		if e := json.Unmarshal(v, &r); e != nil {
			return nil, persistenceError(e)
		}
		if r.CreateKey == key {
			if r.CreateFingerprint != digest {
				return nil, conflict()
			}
			return cloneRawJSON(v), nil
		}
	}
	var r studioRecord
	if e := json.Unmarshal(raw, &r); e != nil {
		return nil, persistenceError(e)
	}
	s.studios[memoryKey(owner, r.Session.ID)] = cloneRawJSON(raw)
	return cloneRawJSON(raw), nil
}
func (s *MySQLStore) ReserveStudioCreation(ctx context.Context, owner, key, digest string, raw json.RawMessage) (json.RawMessage, error) {
	var r studioRecord
	if e := json.Unmarshal(raw, &r); e != nil {
		return nil, persistenceError(e)
	}
	_, e := s.db.ExecContext(ctx, `INSERT INTO novel_studio_sessions (owner_username,studio_id,create_key,create_fingerprint,record_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE create_key=create_key`, owner, r.Session.ID, key, digest, []byte(raw), r.Session.CreatedAt, r.Session.UpdatedAt)
	if e != nil {
		return nil, persistenceError(e)
	}
	var actual string
	var out []byte
	e = s.db.QueryRowContext(ctx, `SELECT create_fingerprint,record_json FROM novel_studio_sessions WHERE owner_username=? AND create_key=? FOR UPDATE`, owner, key).Scan(&actual, &out)
	if e != nil {
		return nil, persistenceError(e)
	}
	if actual != digest {
		return nil, conflict()
	}
	return out, nil
}
func (s *MemoryStore) LoadStudio(_ context.Context, owner, id string, _ bool) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.studios[memoryKey(owner, id)]
	if !ok {
		return nil, studioMissing()
	}
	return cloneRawJSON(r), nil
}
func (s *MemoryStore) SaveStudio(_ context.Context, owner, id string, raw json.RawMessage, create bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.studios[memoryKey(owner, id)]
	if create && ok {
		return conflict()
	}
	if !create && !ok {
		return studioMissing()
	}
	s.studios[memoryKey(owner, id)] = cloneRawJSON(raw)
	return nil
}
func (s *MemoryStore) StudioSessions(_ context.Context, owner string) ([]StudioSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []StudioSession{}
	for k, raw := range s.studios {
		if len(k) < len(owner)+1 || k[:len(owner)+1] != owner+"\x00" {
			continue
		}
		var r studioRecord
		if e := json.Unmarshal(raw, &r); e != nil {
			return nil, persistenceError(e)
		}
		out = append(out, r.Session)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (s *MySQLStore) LoadStudio(ctx context.Context, owner, id string, lock bool) (json.RawMessage, error) {
	q := `SELECT record_json FROM novel_studio_sessions WHERE owner_username=? AND studio_id=?`
	if lock {
		q += " FOR UPDATE"
	}
	var raw []byte
	e := s.db.QueryRowContext(ctx, q, owner, id).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, studioMissing()
	}
	return raw, persistenceError(e)
}
func (s *MySQLStore) SaveStudio(ctx context.Context, owner, id string, raw json.RawMessage, create bool) error {
	var r studioRecord
	if e := json.Unmarshal(raw, &r); e != nil {
		return persistenceError(e)
	}
	var e error
	if create {
		_, e = s.db.ExecContext(ctx, `INSERT INTO novel_studio_sessions (owner_username,studio_id,create_key,create_fingerprint,record_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, owner, id, r.CreateKey, r.CreateFingerprint, []byte(raw), r.Session.CreatedAt, r.Session.UpdatedAt)
	} else {
		_, e = s.db.ExecContext(ctx, `UPDATE novel_studio_sessions SET record_json=?,updated_at=? WHERE owner_username=? AND studio_id=?`, []byte(raw), r.Session.UpdatedAt, owner, id)
	}
	return persistenceError(e)
}
func (s *MySQLStore) StudioSessions(ctx context.Context, owner string) ([]StudioSession, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT record_json FROM novel_studio_sessions WHERE owner_username=? ORDER BY updated_at DESC`, owner)
	if e != nil {
		return nil, persistenceError(e)
	}
	defer rows.Close()
	out := []StudioSession{}
	for rows.Next() {
		var raw []byte
		if e = rows.Scan(&raw); e != nil {
			return nil, persistenceError(e)
		}
		var r studioRecord
		if e = json.Unmarshal(raw, &r); e != nil {
			return nil, persistenceError(e)
		}
		out = append(out, r.Session)
	}
	return out, persistenceError(rows.Err())
}
