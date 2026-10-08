package adminprompt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Lifecycle string

const (
	Draft     Lifecycle = "draft"
	Published Lifecycle = "published"
	Archived  Lifecycle = "archived"
)

type Version struct {
	ID        int64
	Key       string
	Version   int
	Content   string
	Lifecycle Lifecycle
}
type Audit struct {
	Action, Summary, ContentSHA256, RequestID string
	SourceVersion, NewVersion                 int
}
type MemoryStore struct {
	versions map[string][]Version
	audits   []Audit
	auditErr error
}

func NewMemoryStore(values ...Version) *MemoryStore {
	s := &MemoryStore{versions: map[string][]Version{}}
	for _, v := range values {
		s.versions[v.Key] = append(s.versions[v.Key], v)
	}
	return s
}
func (s *MemoryStore) Versions(key string) []Version {
	return append([]Version(nil), s.versions[key]...)
}
func (s *MemoryStore) Audits() []Audit { return append([]Audit(nil), s.audits...) }

type Service struct{ store *MemoryStore }

func NewService(s *MemoryStore) *Service { return &Service{store: s} }
func (s *Service) CreateDraft(_ context.Context, _ int64, key, content, requestID string) (Version, error) {
	vs := s.store.versions[key]
	v := Version{ID: int64(len(vs) + 1), Key: key, Version: len(vs) + 1, Content: content, Lifecycle: Draft}
	s.store.versions[key] = append(vs, v)
	s.store.audits = append(s.store.audits, audit("draft", requestID, 0, v.Version, content))
	return v, nil
}
func (s *Service) ResolveRuntime(_ context.Context, key string) (Version, error) {
	for _, v := range s.store.versions[key] {
		if v.Lifecycle == Published {
			return v, nil
		}
	}
	return Version{}, errors.New("no published prompt")
}
func audit(action, requestID string, source, next int, content string) Audit {
	h := sha256.Sum256([]byte(content))
	return Audit{Action: action, RequestID: requestID, SourceVersion: source, NewVersion: next, ContentSHA256: hex.EncodeToString(h[:])}
}
func (s *Service) Publish(_ context.Context, _ int64, key string, version int, requestID string) error {
	old := append([]Version(nil), s.store.versions[key]...)
	next := append([]Audit(nil), s.store.audits...)
	for i := range s.store.versions[key] {
		if s.store.versions[key][i].Lifecycle == Published {
			s.store.versions[key][i].Lifecycle = Archived
		}
		if s.store.versions[key][i].Version == version {
			s.store.versions[key][i].Lifecycle = Published
		}
	}
	if s.store.auditErr != nil {
		s.store.versions[key] = old
		s.store.audits = next
		return s.store.auditErr
	}
	var v Version
	for _, x := range s.store.versions[key] {
		if x.Version == version {
			v = x
		}
	}
	s.store.audits = append(s.store.audits, audit("publish", requestID, 0, version, v.Content))
	return nil
}
func (s *Service) Restore(ctx context.Context, actor int64, key string, source int, requestID string) (Version, error) {
	var content string
	for _, v := range s.store.versions[key] {
		if v.Version == source {
			content = v.Content
		}
	}
	if content == "" {
		return Version{}, errors.New("not found")
	}
	v, err := s.CreateDraft(ctx, actor, key, content, requestID)
	if err != nil {
		return Version{}, err
	}
	if err = s.Publish(ctx, actor, key, v.Version, requestID); err != nil {
		return Version{}, err
	}
	return s.ResolveRuntime(ctx, key)
}
