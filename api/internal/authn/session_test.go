package authn

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeSessionStore struct {
	mu       sync.Mutex
	users    map[int64]User
	sessions map[string]*SessionRecord
	refresh  map[string]*SessionRecord
}

func newFakeSessionStore(user User) *fakeSessionStore {
	return &fakeSessionStore{
		users:    map[int64]User{user.ID: user},
		sessions: make(map[string]*SessionRecord),
		refresh:  make(map[string]*SessionRecord),
	}
}

func (s *fakeSessionStore) CreateSession(_ context.Context, record SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := record
	s.sessions[record.AccessTokenHash] = &copy
	s.refresh[record.RefreshTokenHash] = &copy
	return nil
}

func (s *fakeSessionStore) ResolveAccess(_ context.Context, hash string, now time.Time) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.sessions[hash]
	if !ok || record.RevokedAt != nil || !now.Before(record.AccessExpiresAt) {
		return User{}, ErrUnauthenticated
	}
	return s.users[record.UserID], nil
}

func (s *fakeSessionStore) RotateByRefresh(_ context.Context, oldHash string, next SessionRecord, now time.Time) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.refresh[oldHash]
	if !ok || record.RevokedAt != nil || !now.Before(record.RefreshExpiresAt) {
		return User{}, ErrUnauthenticated
	}
	user := s.users[record.UserID]
	delete(s.sessions, record.AccessTokenHash)
	delete(s.refresh, record.RefreshTokenHash)
	next.UserID = record.UserID
	copy := next
	s.sessions[next.AccessTokenHash] = &copy
	s.refresh[next.RefreshTokenHash] = &copy
	return user, nil
}

func (s *fakeSessionStore) RevokeByAccess(_ context.Context, hash string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.sessions[hash]
	if !ok {
		return nil
	}
	record.RevokedAt = &now
	return nil
}

func (s *fakeSessionStore) RevokeByRefresh(_ context.Context, hash string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.refresh[hash]
	if !ok {
		return nil
	}
	record.RevokedAt = &now
	return nil
}

func TestManagerAccessRefreshLogoutLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	user := User{ID: 7, Username: "alice", DisplayName: "Alice", Role: "member", TeamID: 3, Capabilities: []string{"batch.view"}}
	store := newFakeSessionStore(user)
	randomBytes := make([]byte, 0, 192)
	for _, value := range []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66} {
		randomBytes = append(randomBytes, bytes.Repeat([]byte{value}, 32)...)
	}
	manager := NewManager(store, Options{
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		Now:        func() time.Time { return now },
		Random:     bytes.NewReader(randomBytes),
	})

	issued, err := manager.Issue(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	if issued.AccessToken == "" || issued.RefreshToken == "" {
		t.Fatal("expected opaque access and refresh credentials")
	}
	if _, exists := store.sessions[issued.AccessToken]; exists {
		t.Fatal("raw access token must never be persisted")
	}
	if _, exists := store.refresh[issued.RefreshToken]; exists {
		t.Fatal("raw refresh token must never be persisted")
	}

	current, err := manager.AuthenticateAccess(context.Background(), issued.AccessToken)
	if err != nil {
		t.Fatalf("authenticate access: %v", err)
	}
	if current.ID != user.ID || current.Username != user.Username {
		t.Fatalf("unexpected current user: %#v", current)
	}

	now = now.Add(2 * time.Minute)
	if _, err := manager.AuthenticateAccess(context.Background(), issued.AccessToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expired access token: got %v, want ErrUnauthenticated", err)
	}

	rotated, refreshedUser, err := manager.Refresh(context.Background(), issued.RefreshToken)
	if err != nil {
		t.Fatalf("refresh session: %v", err)
	}
	if refreshedUser.ID != user.ID {
		t.Fatalf("unexpected refreshed user: %#v", refreshedUser)
	}
	if _, _, err := manager.Refresh(context.Background(), issued.RefreshToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("rotated refresh token reused: got %v, want ErrUnauthenticated", err)
	}
	if _, err := manager.AuthenticateAccess(context.Background(), rotated.AccessToken); err != nil {
		t.Fatalf("authenticate rotated access: %v", err)
	}

	if err := manager.Logout(context.Background(), rotated.AccessToken, rotated.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := manager.AuthenticateAccess(context.Background(), rotated.AccessToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("access after logout: got %v, want ErrUnauthenticated", err)
	}
}

func TestManagerRejectsMissingCredentials(t *testing.T) {
	store := newFakeSessionStore(User{ID: 1, Username: "alice"})
	manager := NewManager(store, Options{})

	if _, err := manager.AuthenticateAccess(context.Background(), ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("empty access: got %v", err)
	}
	if _, _, err := manager.Refresh(context.Background(), ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("empty refresh: got %v", err)
	}
}

func TestManagerConcurrentRefreshAllowsExactlyOneRotation(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	store := newFakeSessionStore(User{ID: 7, Username: "alice"})
	manager := NewManager(store, Options{
		AccessTTL: time.Minute,
		RefreshTTL: time.Hour,
		Now: func() time.Time { return now },
	})
	issued, err := manager.Issue(context.Background(), 7)
	if err != nil { t.Fatalf("issue: %v", err) }

	type result struct {
		credentials Credentials
		err error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			credentials, _, err := manager.Refresh(context.Background(), issued.RefreshToken)
			results <- result{credentials: credentials, err: err}
		}()
	}
	close(start)

	successes := 0
	unauthenticated := 0
	var winner Credentials
	for i := 0; i < 2; i++ {
		item := <-results
		switch {
		case item.err == nil:
			successes++
			winner = item.credentials
		case errors.Is(item.err, ErrUnauthenticated):
			unauthenticated++
		default:
			t.Fatalf("unexpected refresh error: %v", item.err)
		}
	}
	if successes != 1 || unauthenticated != 1 {
		t.Fatalf("successes=%d unauthenticated=%d, want exactly one winner and one rejected stale refresh", successes, unauthenticated)
	}
	if _, err := manager.AuthenticateAccess(context.Background(), winner.AccessToken); err != nil {
		t.Fatalf("winning rotated session must remain valid: %v", err)
	}
}
