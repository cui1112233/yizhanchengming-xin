package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("novel fetch workshop record not found")

type Document struct {
	BookID           string             `json:"bookId"`
	Gender           string             `json:"gender,omitempty"`
	GenderSource     GenderSource       `json:"genderSource,omitempty"`
	GenderReason     string             `json:"genderReason,omitempty"`
	BookInfo         BookInfo           `json:"bookInfo,omitempty"`
	Meta             map[string]any     `json:"meta"`
	BodyRefs         map[string]BodyRef `json:"bodyRefs,omitempty"`
	Original         string             `json:"original,omitempty"`
	OriginalRaw      string             `json:"originalRaw,omitempty"`
	ProcessedText    string             `json:"processedText,omitempty"`
	OriginalRawChars int                `json:"originalRawChars,omitempty"`
	ProcessedChars   int                `json:"processedChars,omitempty"`
	MaxTxt           int                `json:"maxTxt,omitempty"`
	Versions         map[string]any     `json:"versions,omitempty"`
	Logs             []any              `json:"logs"`
	UpdatedAt        string             `json:"updatedAt,omitempty"`
}

type BodyRef struct {
	VersionID    string `json:"versionId"`
	Revision     uint64 `json:"revision"`
	ContentHash  string `json:"contentHash"`
	CharCount    int64  `json:"charCount"`
	State        string `json:"state"`
	UpdatedAt    string `json:"updatedAt"`
	LastNeededAt string `json:"lastNeededAt,omitempty"`
	ReleasableAt string `json:"releasableAt,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
}

type BodyRecord struct {
	BodyRef
	BookID  string `json:"bookId"`
	Content string `json:"content"`
}

// RunRecord is the credential-free execution audit record for one book stage.
// Model credentials stay in Go's account-scoped catalog and are intentionally
// not represented in the novel-fetch audit record.
type RunRecord struct {
	RunID            string `json:"runId"`
	BookID           string `json:"bookId"`
	Stage            string `json:"stage"`
	Status           string `json:"status"`
	Attempts         int    `json:"attempts"`
	TextModelID      string `json:"textModelId"`
	ModelID          string `json:"modelId"`
	ModelDisplayName string `json:"modelDisplayName"`
	RequestID        string `json:"requestId,omitempty"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	StartedAt        string `json:"startedAt,omitempty"`
	FinishedAt       string `json:"finishedAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

type DeleteResult struct {
	Requested int            `json:"requested"`
	Deleted   int            `json:"deleted"`
	Results   []DeleteRecord `json:"results"`
}

type DeleteRecord struct {
	BookID  string `json:"bookId"`
	Deleted bool   `json:"deleted"`
}

type BodyStore interface {
	PutBody(context.Context, string, BodyRecord) (BodyRef, error)
	GetBody(context.Context, string, string, string) (BodyRecord, error)
	ListBodyRefs(context.Context, string, string) ([]BodyRef, error)
	DeleteBody(context.Context, string, string, string) (bool, error)
}

type Store interface {
	BodyStore
	GetDocument(context.Context, string, string) (Document, error)
	PutDocument(context.Context, string, Document) error
	ListDocuments(context.Context, string) ([]Document, error)
	DeleteDocuments(context.Context, string, []string) (DeleteResult, error)
	GetConfig(context.Context, string) (map[string]any, error)
	PutConfig(context.Context, string, map[string]any) error
	PutRun(context.Context, string, RunRecord) error
	ListRuns(context.Context, string, string) ([]RunRecord, error)
}

func normalizeDocument(document Document) Document {
	document.BookID = strings.TrimSpace(document.BookID)
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	if document.BodyRefs == nil {
		document.BodyRefs = map[string]BodyRef{}
	}
	if document.Versions == nil {
		document.Versions = map[string]any{}
	}
	if document.Logs == nil {
		document.Logs = []any{}
	}
	return document
}

func cloneDocument(document Document) Document {
	raw, _ := json.Marshal(document)
	var cloned Document
	_ = json.Unmarshal(raw, &cloned)
	return normalizeDocument(cloned)
}

func cloneMap(value map[string]any) map[string]any {
	raw, _ := json.Marshal(value)
	var cloned map[string]any
	_ = json.Unmarshal(raw, &cloned)
	if cloned == nil {
		cloned = map[string]any{}
	}
	return cloned
}

func cloneBodyRecord(body BodyRecord) BodyRecord {
	return body
}

func normalizeRunRecord(record RunRecord) RunRecord {
	record.RunID = strings.TrimSpace(record.RunID)
	record.BookID = strings.TrimSpace(record.BookID)
	record.Stage = strings.TrimSpace(record.Stage)
	record.Status = strings.TrimSpace(record.Status)
	record.TextModelID = strings.TrimSpace(record.TextModelID)
	record.ModelID = strings.TrimSpace(record.ModelID)
	record.ModelDisplayName = strings.TrimSpace(record.ModelDisplayName)
	record.RequestID = strings.TrimSpace(record.RequestID)
	record.ErrorMessage = strings.TrimSpace(record.ErrorMessage)
	return record
}

type MemoryStore struct {
	mu      sync.RWMutex
	docs    map[string]map[string]Document
	configs map[string]map[string]any
	bodies  map[string]map[string]map[string]BodyRecord
	runs    map[string]map[string]RunRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		docs:    map[string]map[string]Document{},
		configs: map[string]map[string]any{},
		bodies:  map[string]map[string]map[string]BodyRecord{},
		runs:    map[string]map[string]RunRecord{},
	}
}

func (s *MemoryStore) PutRun(_ context.Context, owner string, record RunRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record = normalizeRunRecord(record)
	if record.RunID == "" || record.BookID == "" || record.Stage == "" || record.Status == "" || record.TextModelID == "" || record.ModelID == "" {
		return errors.New("run audit fields are required")
	}
	if s.runs[owner] == nil {
		s.runs[owner] = map[string]RunRecord{}
	}
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	s.runs[owner][record.RunID] = record
	return nil
}

func (s *MemoryStore) ListRuns(_ context.Context, owner, bookID string) ([]RunRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]RunRecord, 0)
	for _, record := range s.runs[owner] {
		if bookID == "" || record.BookID == bookID {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt == result[j].UpdatedAt {
			return result[i].RunID < result[j].RunID
		}
		return result[i].UpdatedAt > result[j].UpdatedAt
	})
	return result, nil
}

func (s *MemoryStore) GetDocument(_ context.Context, owner, bookID string) (Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	document, ok := s.docs[owner][bookID]
	if !ok {
		return Document{}, ErrNotFound
	}
	return cloneDocument(document), nil
}

func (s *MemoryStore) PutDocument(_ context.Context, owner string, document Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	document = normalizeDocument(document)
	document.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if s.docs[owner] == nil {
		s.docs[owner] = map[string]Document{}
	}
	s.docs[owner][document.BookID] = cloneDocument(document)
	return nil
}

func (s *MemoryStore) ListDocuments(_ context.Context, owner string) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Document, 0, len(s.docs[owner]))
	for _, document := range s.docs[owner] {
		result = append(result, cloneDocument(document))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt == result[j].UpdatedAt {
			return result[i].BookID < result[j].BookID
		}
		return result[i].UpdatedAt > result[j].UpdatedAt
	})
	return result, nil
}

func (s *MemoryStore) DeleteDocuments(_ context.Context, owner string, ids []string) (DeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := DeleteResult{Requested: len(ids), Results: make([]DeleteRecord, 0, len(ids))}
	for _, id := range ids {
		_, exists := s.docs[owner][id]
		if exists {
			delete(s.docs[owner], id)
			result.Deleted++
		}
		result.Results = append(result.Results, DeleteRecord{BookID: id, Deleted: exists})
	}
	return result, nil
}

func (s *MemoryStore) GetConfig(_ context.Context, owner string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return EnsureWorkshopDefaults(cloneMap(s.configs[owner])), nil
}

func (s *MemoryStore) PutConfig(_ context.Context, owner string, settings map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[owner] = cloneMap(settings)
	return nil
}

func (s *MemoryStore) PutBody(_ context.Context, owner string, body BodyRecord) (BodyRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	body.BookID = strings.TrimSpace(body.BookID)
	body.VersionID = strings.TrimSpace(body.VersionID)
	if body.State == "" {
		body.State = "ready"
	}
	_, hash, chars, err := encodeBody(body.Content)
	if err != nil {
		return BodyRef{}, err
	}
	if s.bodies[owner] == nil {
		s.bodies[owner] = map[string]map[string]BodyRecord{}
	}
	if s.bodies[owner][body.BookID] == nil {
		s.bodies[owner][body.BookID] = map[string]BodyRecord{}
	}
	if previous, ok := s.bodies[owner][body.BookID][body.VersionID]; ok {
		body.Revision = previous.Revision + 1
	} else {
		body.Revision = 1
	}
	body.ContentHash = hash
	body.CharCount = chars
	now := time.Now().UTC().Format(time.RFC3339Nano)
	body.UpdatedAt = now
	body.LastNeededAt = now
	body.ReleasableAt = ""
	body.ExpiresAt = ""
	s.bodies[owner][body.BookID][body.VersionID] = cloneBodyRecord(body)
	return body.BodyRef, nil
}

func (s *MemoryStore) GetBody(_ context.Context, owner, bookID, versionID string) (BodyRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, ok := s.bodies[owner][bookID][versionID]
	if !ok {
		return BodyRecord{}, ErrNotFound
	}
	body.LastNeededAt = time.Now().UTC().Format(time.RFC3339Nano)
	s.bodies[owner][bookID][versionID] = body
	return cloneBodyRecord(body), nil
}

func (s *MemoryStore) ListBodyRefs(_ context.Context, owner, bookID string) ([]BodyRef, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bodies := s.bodies[owner][bookID]
	refs := make([]BodyRef, 0, len(bodies))
	for _, body := range bodies {
		refs = append(refs, body.BodyRef)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].VersionID < refs[j].VersionID })
	return refs, nil
}

func (s *MemoryStore) DeleteBody(_ context.Context, owner, bookID, versionID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions := s.bodies[owner][bookID]
	if versions == nil {
		return false, nil
	}
	if _, ok := versions[versionID]; !ok {
		return false, nil
	}
	delete(versions, versionID)
	if document, ok := s.docs[owner][bookID]; ok {
		document = normalizeDocument(document)
		delete(document.BodyRefs, versionID)
		s.docs[owner][bookID] = cloneDocument(document)
	}
	return true, nil
}
